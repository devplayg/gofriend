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

	"github.com/devplayg/golibs/orm"
	_ "github.com/mattn/go-sqlite3"

	"github.com/devplayg/gofriend"
)

const (
	YYYYMMDDHH24MISS = "2006-01-02 15:04:05"
)
const (
	FILE_ADDED = iota
	FILE_MODIFIED
	FILE_DELETED
)

type Backup struct {
	srcDir       string
	dstDir       string
	db           *sql.DB
	dbOriginFile string
	dbLogFile    string
	dbFile       string
	tempDir      string
	oOrigin      orm.Ormer
	oLog         orm.Ormer
	t            time.Time
}

type Summary struct {
	ID            int64
	Date          time.Time
	SrcDir        string
	DstDir        string
	State         int
	TotalSize     uint64
	TotalCount    uint32
	BackupNew     uint32
	BackupDeleted uint32
	BackupSuccess uint32
	BackupFailure uint32
	BackupSize    uint64
	ExecutionTime float64
	Message       string
}

func newSummary(Date time.Time, SrcDir string) *Summary {
	return &Summary{
		Date:   Date,
		SrcDir: SrcDir,
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

type FileMap map[string]*File

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
	b.t = time.Now()

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

	err = orm.RegisterDataBase("orgin", "sqlite3", b.dbOriginFile)
	gofriend.CheckErr(err)

	err = orm.RegisterDataBase("default", "sqlite3", b.dbLogFile)
	gofriend.CheckErr(err)

	b.oOrigin = orm.NewOrm()
	b.oOrigin.Using("origin")

	b.oLog = orm.NewOrm()
	b.oLog.Using("default")

	//	b.oOrigin = orm.
	//	b.oOrigin, _ = orm.GetDB("origin")
	//	b.oOrigin.Using("origin")
	//	b.oLog, _ = orm.GetDB("default")
	//	b.oLog.Using("default")

	_, err = b.oOrigin.Raw(`
		CREATE TABLE IF NOT EXISTS bak_origin (
			path text not null,
			size int not null,
			mtime text not null
		);
	`).Exec()

	b.oLog.Using("default")
	_, err = b.oLog.Raw(`
		CREATE TABLE IF NOT EXISTS bak_summary (
			id integer not null primary key autoincrement,
			date integer not null  DEFAULT CURRENT_TIMESTAMP,
			src_dir text not null default '',
			dst_dir text not null default '',
			state integer not null default 0,
			total_size integer not null default 0,
			total_count integer not null default 0,
			backup_new integer not null default 0,
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
	`).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (b *Backup) getLastSummay() *Summary {
	var summary Summary
	err := b.oLog.Raw(`
			select id, date, src_dir, dst_dir, state, total_size, total_count, message
			from bak_summary
			where id = (select max(id) from bak_summary)
		`).QueryRow(&summary)
	gofriend.CheckErr(err)
	return &summary
	return nil
}

func (b *Backup) getBackupLog(id int64) FileMap {
	fm := make(map[string]*File, 0)

	//	var files []File
	//	num, err := o.Raw("SELECT id, name FROM user WHERE id = ?", 1).QueryRows(&users)
	//	if err == nil {
	//		fmt.Println("user nums: ", num)
	//	}
	return fm
}

func (b *Backup) getLastBackupLog() FileMap {
	var fm FileMap

	summary := b.getLastSummay()
	if summary != nil {
		fm = b.getBackupLog(summary.ID)
	}

	return fm
}

func (b *Backup) Start() error {
	t1 := time.Now()
	oldMap := b.getLastBackupLog()
	newMap := make(map[string]*File, 10)

	wg := new(sync.WaitGroup)
	summary := newSummary(b.t, b.srcDir)
	summary.State = 1

	// Backup
	err := filepath.Walk(b.srcDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			wg.Add(1)

			go func(path string, f os.FileInfo) {
				atomic.AddUint32(&summary.TotalCount, 1)
				atomic.AddUint64(&summary.TotalSize, uint64(f.Size()))
				fi := newFile(path, f.Size(), f.ModTime())

				if oldFile, ok := oldMap[path]; ok {
					if oldFile.ModTime != f.ModTime() || oldFile.Size != f.Size() {
						fi.State = FILE_MODIFIED
						backupPath, err := b.BackupFile(path)
						if err != nil {
							atomic.AddUint32(&summary.BackupFailure, 1)
							gofriend.CheckErr(err)
						} else {
							atomic.AddUint32(&summary.BackupSuccess, 1)
							atomic.AddUint64(&summary.BackupSize, uint64(f.Size()))
							os.Chtimes(backupPath, f.ModTime(), f.ModTime())
						}
					}

					delete(oldMap, path)

				} else {
					//					if b.t.Sub(f.ModTime()).Seconds() < 86400 {
					fi.State = FILE_ADDED
					atomic.AddUint32(&summary.BackupNew, 1)
					backupPath, err := b.BackupFile(path)
					if err != nil {
						atomic.AddUint32(&summary.BackupFailure, 1)
						gofriend.CheckErr(err)
					} else {
						atomic.AddUint32(&summary.BackupSuccess, 1)
						atomic.AddUint64(&summary.BackupSize, uint64(f.Size()))
						os.Chtimes(backupPath, f.ModTime(), f.ModTime())
					}
					//					}
				}
				newMap[path] = fi

				// Done
				wg.Done()
			}(path, f)
		}
		return nil
	})
	gofriend.CheckErr(err)
	wg.Wait()

	// Rename directory
	lastDir := filepath.Join(b.dstDir, b.t.Format("20060102"))
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
	log.Printf("Backup time: %3.1f\n", time.Since(t1).Seconds())

	// Write log
	t1 = time.Now()
	summary.ExecutionTime = time.Since(b.t).Seconds()
	b.writeToDatabase(summary, newMap, oldMap)
	log.Printf("Logging time: %3.1f\n", time.Since(t1).Seconds())

	return err
}

func (b *Backup) writeToDatabase(s *Summary, nm map[string]*File, om map[string]*File) error {
	res, err := b.oLog.Raw(`
		insert into bak_summary(date,src_dir,dst_dir,state,total_size,total_count,backup_new,backup_deleted,backup_success,backup_failure,backup_size,execution_time,message)
			values(?,?,?,?,?,?,?,?,?,?,?,?,?)
	`,
		s.Date.Format(YYYYMMDDHH24MISS),
		s.SrcDir,
		s.DstDir,
		s.State,
		s.TotalSize,
		s.TotalCount,
		s.BackupNew,
		s.BackupDeleted,
		s.BackupSuccess,
		s.BackupFailure,
		s.BackupSize,
		s.ExecutionTime,
		s.Message,
	).Exec()
	gofriend.CheckErr(err)
	if err == nil {
		s.ID, _ = res.LastInsertId()
	} else {
		return err
	}

	var lines []string

	// Modified or added files
	i := 0
	for _, f := range nm {
		if f.State > 0 {
			lines = append(lines, fmt.Sprintf("select %d, '%s', %d, '%s', %d, '%s'", s.ID, f.Path, f.Size, f.ModTime.Format(YYYYMMDDHH24MISS), f.State, f.Message))
			i += 1
		}

		if i%3 == 0 || len(nm) == i {
			query := fmt.Sprintf("insert into bak_log(id, path, size, mtime, state, message) %s", strings.Join(lines, " union all "))
			b.oLog.Raw(query).Exec()
			lines = nil
		}

	}

	// Deleted files
	i = 0
	lines = nil
	for _, f := range om {
		f.State = FILE_DELETED
		lines = append(lines, fmt.Sprintf("select %d, '%s', %d, '%s', %d, '%s'", s.ID, f.Path, f.Size, f.ModTime.Format(YYYYMMDDHH24MISS), f.State, f.Message))
		i += 1

		if i%3 == 0 || len(nm) == i {
			query := fmt.Sprintf("insert into bak_log(id, path, size, mtime, state, message) %s", strings.Join(lines, " union all "))
			b.oLog.Raw(query).Exec()
			lines = nil
		}
	}

	// All files
	i = 0
	lines = nil
	for _, f := range om {
		f.State = FILE_DELETED
		lines = append(lines, fmt.Sprintf("select %d, '%s', %d, '%s', %d, '%s'", s.ID, f.Path, f.Size, f.ModTime.Format(YYYYMMDDHH24MISS), f.State, f.Message))
		i += 1

		if i%3 == 0 || len(nm) == i {
			query := fmt.Sprintf("insert into bak_origin(path, size, mtime) %s", strings.Join(lines, " union all "))
			b.oOrigin.Raw(query).Exec()
			lines = nil
		}
	}

	return nil
}

func (b *Backup) Close() error {
	b.db.Close()
	return nil
}

func (b *Backup) BackupFile(path string) (string, error) {
	log.Printf("Backup: %s\n", path)
	// Set source
	from, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer from.Close()

	// Set destination
	dst := b.tempDir + path[len(b.srcDir):]
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
