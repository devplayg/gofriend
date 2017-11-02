package backup

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	FileModified = 1 << iota // 1
	FileAdded    = 1 << iota // 2
	FileDeleted  = 1 << iota // 4
)

type Backup struct {
	srcDir       string
	dstDir       string
	dbOriginFile string
	dbLogFile    string
	dbFile       string
	tempDir      string
	dbOrigin     *sql.DB
	dbLog        *sql.DB
}

type Summary struct {
	ID         int64
	Date       time.Time
	SrcDir     string
	DstDir     string
	State      int
	TotalSize  uint64
	TotalCount uint32

	BackupAdded    uint32
	BackupModified uint32
	BackupDeleted  uint32

	BackupSuccess uint32
	BackupFailure uint32

	BackupSize    uint64
	ExecutionTime float64
	Message       string

	BackupTime  float64
	LoggingTime float64
}

func newSummary(SrcDir string) *Summary {
	return &Summary{
		Date:   time.Now(),
		SrcDir: SrcDir,
		State:  1,
	}
}

type File struct {
	Path    string
	Size    int64
	ModTime time.Time
	Result  int
	State   int
	Message string
}

func newFile(path string, size int64, modTime time.Time) *File {
	return &File{
		Path:    path,
		Size:    size,
		ModTime: modTime,
	}

}

//type FileMap map[string]*File

func NewBackup(srcDir, dstDir string) *Backup {
	b := Backup{
		srcDir:       filepath.Clean(srcDir),
		dstDir:       filepath.Clean(dstDir),
		dbOriginFile: filepath.Join(filepath.Clean(dstDir), "backup_origin.db"),
		dbLogFile:    filepath.Join(filepath.Clean(dstDir), "backup_log.db"),
	}
	return &b
}

// Initialize
func (b *Backup) Initialize() error {
	var err error
	err = b.initDir()
	if err != nil {
		return err
	}
	err = b.initDB()
	if err != nil {
		return err
	}

	return nil
}

// Initialize directories
func (b *Backup) initDir() error {
	if _, err := os.Stat(b.srcDir); os.IsNotExist(err) {
		return err
	}

	if _, err := os.Stat(b.dstDir); os.IsNotExist(err) {
		return err
	}

	tempDir, err := ioutil.TempDir(b.dstDir, "bak")
	if err != nil {
		return err
	}
	b.tempDir = tempDir

	return nil
}

// Initialize database
func (b *Backup) initDB() error {
	var err error
	var query string

	// Set databases
	b.dbOrigin, err = sql.Open("sqlite3", b.dbOriginFile)
	if err != nil {
		return err
	}
	b.dbLog, err = sql.Open("sqlite3", b.dbLogFile)
	if err != nil {
		return err
	}

	// Original database
	query = `
		CREATE TABLE IF NOT EXISTS bak_origin (
			path text not null,
			size int not null,
			mtime text not null
		);
	`
	_, err = b.dbOrigin.Exec(query)
	if err != nil {
		return err
	}

	// Log database
	query = `
		CREATE TABLE IF NOT EXISTS bak_summary (
			id integer not null primary key autoincrement,
			date integer not null  DEFAULT CURRENT_TIMESTAMP,
			src_dir text not null default '',
			dst_dir text not null default '',
			state integer not null default 0,
			total_size integer not null default 0,
			total_count integer not null default 0,
			backup_modified integer not null default 0,
			backup_added integer not null default 0,
			backup_deleted integer not null default 0,
			backup_success integer not null default 0,
			backup_failure integer not null default 0,
			backup_size integer not null default 0,
			execution_time real not null default 0.0,
			message text not null default ''
		);
		CREATE INDEX IF NOT EXISTS ix_bak_summary ON bak_summary(date);
		CREATE TABLE IF NOT EXISTS bak_log(
			id int not null,
			path text not null,
			size int not null,
			mtime text not null,
			state int not null,
			message text not null
		);
		CREATE INDEX IF NOT EXISTS ix_bak_log_id on bak_log(id);
	`
	_, err = b.dbLog.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backup) getOriginMap() (sync.Map, int) {
	m := sync.Map{}
	rows, err := b.dbOrigin.Query("select path, size, mtime from bak_origin")
	checkErr(err)

	var count = 0
	var path string
	var size int64
	var modTime string
	for rows.Next() {
		f := newFile("", 0, time.Now())
		err = rows.Scan(&path, &size, &modTime)
		checkErr(err)
		f.Path = path
		f.Size = size
		f.ModTime, _ = time.Parse(time.RFC3339, modTime)
		m.Store(path, f)
		count += 1
	}
	return m, count
}

func (b *Backup) Start() (*Summary, error) {

	// Create summary
	summary := newSummary(b.srcDir)
	lastSummary := b.getLastSummary()

	// Create maps
	originMap, originCount := b.getOriginMap()
	newMap := sync.Map{}

	// Insert initial data
	if originCount < 1 || b.srcDir != lastSummary.SrcDir {
		summary.State = 2
		log.Println("Inserting initial data..")
		err := filepath.Walk(b.srcDir, func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() {
				//				path = strings.Replace(path, "'", "\'\'", -1)
				fi := newFile(path, f.Size(), f.ModTime())
				newMap.Store(path, fi)
				summary.TotalCount += 1
			}
			return nil
		})
		checkErr(err)
		os.RemoveAll(b.tempDir)

		b.writeToDatabase(summary, newMap, sync.Map{})
		summary.LoggingTime = time.Since(summary.Date).Seconds()
		log.Printf("Time: %3.1fs\n", summary.LoggingTime)
		return nil, nil
	}

	// Backup
	log.Printf("Reading all files in %s\n", b.srcDir)
	summary.State = 3
	wg := new(sync.WaitGroup)
	err := filepath.Walk(b.srcDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			wg.Add(1)

			go func(path string, f os.FileInfo) {

				atomic.AddUint32(&summary.TotalCount, 1)
				atomic.AddUint64(&summary.TotalSize, uint64(f.Size()))
				fi := newFile(path, f.Size(), f.ModTime())

				if inf, ok := originMap.Load(path); ok {
					last := inf.(*File)

					//					if !last.ModTime.Equal(f.ModTime()) {
					//					log.Printf("%s = %s, %b\n", last.ModTime.Format(time.RFC3339Nano), f.ModTime().Format(time.RFC3339Nano), last.ModTime.Equal(f.ModTime()))
					//					log.Printf("%d = %d, %b\n", last.ModTime.Unix(), f.ModTime().Unix(), last.ModTime.Unix() == f.ModTime().Unix())

					//						log.Printf("%d = %d\n", last.Size, f.Size())
					//					log.Println(last.ModTime.Equal(f.ModTime()))

					//					}

					if last.ModTime.Unix() != f.ModTime().Unix() || last.Size != f.Size() {
						fi.State = FileModified
						atomic.AddUint32(&summary.BackupModified, 1)
						backupPath, err := b.BackupFile(path)
						if err != nil {
							atomic.AddUint32(&summary.BackupFailure, 1)
							checkErr(err)
							fi.Message = err.Error()
						} else {
							atomic.AddUint32(&summary.BackupSuccess, 1)
							atomic.AddUint64(&summary.BackupSize, uint64(f.Size()))
							os.Chtimes(backupPath, f.ModTime(), f.ModTime())
						}
					}
					originMap.Delete(path)
				} else {
					fi.State = FileAdded
					atomic.AddUint32(&summary.BackupAdded, 1)
					backupPath, err := b.BackupFile(path)
					if err != nil {
						atomic.AddUint32(&summary.BackupFailure, 1)
						checkErr(err)
						fi.Message = err.Error()
					} else {
						atomic.AddUint32(&summary.BackupSuccess, 1)
						atomic.AddUint64(&summary.BackupSize, uint64(f.Size()))
						os.Chtimes(backupPath, f.ModTime(), f.ModTime())
					}
				}
				newMap.Store(path, fi)

				// Done
				wg.Done()
			}(path, f)
		}
		return nil
	})
	checkErr(err)
	wg.Wait()

	// Rename directory
	lastDir := filepath.Join(b.dstDir, summary.Date.Format("20060102"))
	err = os.Rename(b.tempDir, lastDir)

	if err == nil {
		summary.DstDir = lastDir
	} else {

		i := 1
		for err != nil && i <= 10 {
			altDir := lastDir + "_" + strconv.Itoa(i)
			err = os.Rename(b.tempDir, altDir)
			if err == nil {
				summary.DstDir = altDir
			}
			i += 1
		}
		if err != nil {
			summary.Message = err.Error()
			summary.State = -1
			summary.DstDir = b.tempDir
			os.RemoveAll(b.tempDir)
		}
	}
	summary.BackupTime = time.Since(summary.Date).Seconds() - summary.LoggingTime

	// Write log
	err = b.writeToDatabase(summary, newMap, originMap)
	return summary, err
}

func (b *Backup) getLastSummary() *Summary {
	rows, _ := b.dbLog.Query(`
		select src_dir
		from bak_summary
		where id = (select max(id) from bak_summary)
	`)
	defer rows.Close()
	var srcDir string
	rows.Next()
	rows.Scan(&srcDir)

	s := newSummary(srcDir)
	return s
}

func (b *Backup) writeToDatabase(s *Summary, newMap sync.Map, originMap sync.Map) error {
	t := time.Now()
	stmt, _ := b.dbLog.Prepare(`
		insert into bak_summary(date,src_dir,dst_dir,state,total_size,total_count,backup_modified,backup_added,backup_deleted,backup_success,backup_failure,backup_size,execution_time,message)
		values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	res, err := stmt.Exec(
		s.Date.Format(time.RFC3339),
		s.SrcDir,
		s.DstDir,
		s.State,
		s.TotalSize,
		s.TotalCount,
		s.BackupModified,
		s.BackupAdded,
		s.BackupDeleted,
		s.BackupSuccess,
		s.BackupFailure,
		s.BackupSize,
		s.ExecutionTime,
		s.Message,
	)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()

	var maxInsertSize uint32 = 500
	var lines []string
	var eventLines []string
	var i uint32 = 0
	var j uint32 = 0

	// Delete original data
	query := "delete from bak_origin"
	stmt, _ = b.dbOrigin.Prepare(query)
	_, err = stmt.Exec()

	// Modified or added files
	newMap.Range(func(key, value interface{}) bool {
		f := value.(*File)
		path := strings.Replace(f.Path, "'", "''", -1)
		lines = append(lines, fmt.Sprintf("select '%s', %d, '%s'", path, f.Size, f.ModTime.Format(time.RFC3339)))
		i += 1

		if i%maxInsertSize == 0 || i == s.TotalCount {
			err := b.insertIntoOrigin(lines)
			checkErr(err)
			lines = nil
		}

		if f.State > 0 {
			eventLines = append(eventLines, fmt.Sprintf("select %d, '%s', %d, '%s', %d, '%s'", s.ID, path, f.Size, f.ModTime.Format(time.RFC3339), f.State, f.Message))
			j += 1

			if j%maxInsertSize == 0 {
				err := b.insertIntoLog(eventLines)
				checkErr(err)
				eventLines = nil
			}
		}
		return true
	})
	if len(eventLines) > 0 {
		err := b.insertIntoLog(eventLines)
		checkErr(err)
		eventLines = nil
	}

	// Deleted files
	eventLines = nil
	j = 0
	originMap.Range(func(key, value interface{}) bool {
		f := value.(*File)
		f.State = FileDeleted
		path := strings.Replace(f.Path, "'", "''", -1)
		eventLines = append(eventLines, fmt.Sprintf("select %d, '%s', %d, '%s', %d, '%s'", s.ID, path, f.Size, f.ModTime.Format(time.RFC3339), f.State, f.Message))
		j += 1

		if j%maxInsertSize == 0 {
			err := b.insertIntoLog(eventLines)
			checkErr(err)
			eventLines = nil
		}
		return true
	})
	if len(eventLines) > 0 {
		err := b.insertIntoLog(eventLines)
		checkErr(err)
		eventLines = nil
	}
	atomic.AddUint32(&s.BackupDeleted, j)

	s.LoggingTime = time.Since(t).Seconds()
	msg := fmt.Sprintf("Backup time: %3.1f, Logging time: %3.1fs", s.BackupTime, s.LoggingTime)
	if s.State == 2 {
		msg = "Data initialized, " + msg
	}

	stmt, _ = b.dbLog.Prepare("update bak_summary set backup_deleted = ?, message = ?, execution_time = ? where id = ?")
	stmt.Exec(s.BackupDeleted, msg, s.BackupTime+s.LoggingTime, s.ID)

	return nil
}

func (b *Backup) insertIntoLog(rows []string) error {
	query := fmt.Sprintf("insert into bak_log(id, path, size, mtime, state, message) %s", strings.Join(rows, " union all "))
	stmt, _ := b.dbLog.Prepare(query)
	_, err := stmt.Exec()
	return err
}

func (b *Backup) insertIntoOrigin(rows []string) error {
	query := fmt.Sprintf("insert into bak_origin(path, size, mtime) %s", strings.Join(rows, " union all "))
	stmt, err := b.dbOrigin.Prepare(query)
	defer func() {
		if r := recover(); r != nil {
			if err != nil {
				log.Println(query)
			}
		}
	}()
	checkErr(err)
	_, err = stmt.Exec()
	return err
}

func (b *Backup) Close() error {
	b.dbOrigin.Close()
	b.dbLog.Close()
	return nil
}

func (b *Backup) BackupFile(path string) (string, error) {
	// Set source
	from, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer from.Close()

	// Set destination
	dst := filepath.Join(b.tempDir, path[len(b.srcDir):])
	err = os.MkdirAll(filepath.Dir(dst), 0644)
	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer to.Close()

	// Copy
	_, err = io.Copy(to, from)
	if err != nil {
		return "", err
	}

	return dst, err
}

func checkErr(err error) {
	if err != nil {
		log.Printf("[Error] %s\n", err.Error())
	}
}
