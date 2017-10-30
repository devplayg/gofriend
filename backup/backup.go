package backup

import (
	"database/sql"

	"github.com/devplayg/golibs/orm"
	_ "github.com/mattn/go-sqlite3"
	//"io"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/devplayg/gofriend"
)

const (
	YYYYMMDDHH24MISS = "2006-01-02 15:04:05"
)

type Backup struct {
	srcDir  string
	dstDir  string
	db      *sql.DB
	dbFile  string
	tempDir string
	o       orm.Ormer
	t       time.Time
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

func NewBackup(srcDir, dstDir, db string) *Backup {
	b := Backup{
		srcDir: filepath.Clean(srcDir),
		dstDir: filepath.Clean(dstDir),
		dbFile: db,
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
	err := orm.RegisterDataBase("default", "sqlite3", b.dbFile)
	gofriend.CheckErr(err)
	b.o = orm.NewOrm()
	b.o.Using("default")

	_, err = b.o.Raw(`
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

		CREATE TABLE IF NOT EXISTS bak_backup(
			id int not null,
			path text not null,
			size int not null,
			mtime text not null,
			state int not null,
			message text not null
		);

		CREATE INDEX IF NOT EXISTS ix_bak_backup_id on bak_backup(id);
		CREATE INDEX IF NOT EXISTS ix_bak_backup_id_path on bak_backup(id, path);

	`).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (b *Backup) getLastSummay() *Summary {
	var summary Summary
	o := orm.NewOrm()
	err := o.Raw(`
		select id, date, src_dir, dst_dir, state, total_size, total_count, message
		from bak_summary
		where id = (select max(id) from bak_summary)
	`).QueryRow(&summary)
	gofriend.CheckErr(err)
	return &summary
}

func (b *Backup) getBackupLog(id int64) FileMap {
	fm := make(map[string]*File, 0)

	//var files []File
	//num, err := o.Raw("SELECT id, name FROM user WHERE id = ?", 1).QueryRows(&users)
	//if err == nil {
	//	fmt.Println("user nums: ", num)
	//}
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
	//	t1 := time.Now()
	oldMap := b.getLastBackupLog()
	newMap := make(map[string]*File, 10)
	wg := new(sync.WaitGroup)
	summary := newSummary(b.t, b.srcDir)
	summary.State = 1

	// Backup
	err := filepath.Walk(b.srcDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			wg.Add(1)

			atomic.AddUint32(&summary.TotalCount, 1)
			atomic.AddUint64(&summary.TotalSize, uint64(f.Size()))

			go func(path string, f os.FileInfo) {
				newMap[path] = newFile(path, f.Size(), f.ModTime())

				if oldFile, ok := oldMap[path]; ok {
					if oldFile.ModTime != f.ModTime() || oldFile.Size != f.Size() {
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

				} else {
					if b.t.Sub(f.ModTime()).Seconds() < 86400 {
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
					}
				}

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
	//	log.Printf("Backup time: %3.1f\n", time.Since(t1).Seconds())

	//	// Write log
	//	t1 = time.Now()
	//	summary.ExecutionTime = time.Since(b.t).Seconds()
	//	b.writeLog(summary, newMap)
	//	log.Printf("Logging time: %3.1f\n", time.Since(t1).Seconds())

	return err
}

func (b *Backup) writeLog(s *Summary, m map[string]*File) error {
	res, err := b.o.Raw(`
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

	pstmt, err := b.o.Raw("insert into bak_backup(id, path, size, mtime, state, message) values(?,?,?,?,?,?);").Prepare()
	for _, f := range m {
		pstmt.Exec(s.ID, f.Path, f.Size, f.ModTime.Format(YYYYMMDDHH24MISS), f.State, f.Message)
	}
	pstmt.Close()

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
