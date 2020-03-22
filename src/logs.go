package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const DateFormat = "2006-01-02" // As the unit of time in days to split the logs.

type LEVEL byte

const (
	TRACE LEVEL = iota
	INFO
	WARN
	ERROR
	OFF
)

var fileLog *FileLogger

type LoggerConf struct {
	FileDir  string
	FileName string
	Prefix   string
	Level    string
}

type FileLogger struct {
	fileDir  string
	fileName string
	prefix   string
	date     *time.Time
	logFile  *os.File
	lg       *log.Logger
	logLevel LEVEL
	mu       *sync.RWMutex
	logChan  chan string
}

var Logs = new(FileLogger)

// Initialize the logging configuration.
func BootLogger() (err error) {
	conf := &LoggerConf{
		FileDir:  Configure.Section("logs").Get("fileDir"),  // Get log dir setting.
		FileName: Configure.Section("logs").Get("filename"), // Get log filename setting.
		Prefix:   Configure.Section("logs").Get("prefix"),   // Get log filename prefix setting.
		Level:    Configure.Section("logs").Get("level"),    // Get log level setting.
	}

	f := &FileLogger{
		fileDir:  conf.FileDir,
		fileName: conf.FileName,
		prefix:   conf.Prefix,
		mu:       new(sync.RWMutex),
		logChan:  make(chan string, 5000),
	}

	// Open close and set the logging level.
	switch conf.Level {
	case "off":
		f.logLevel = OFF // Setting off a log.

	case "trace":
		f.logLevel = TRACE // Setting trace level.

	case "warn":
		f.logLevel = WARN // Setting warn level.

	case "error":
		f.logLevel = ERROR // Setting error level.

	default:
		f.logLevel = INFO // Setting info level.
	}

	t, _ := time.Parse(DateFormat, time.Now().Format(DateFormat))
	f.date = &t

	if f.isMustSplit() {
		if err = f.split(); err != nil {
			return
		}

	} else {
		f.isExistOrCreate()

		logFile := filepath.Join(f.fileDir, f.fileName)

		f.logFile, err = os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return
		}

		f.lg = log.New(f.logFile, f.prefix, log.LstdFlags|log.Lmicroseconds)
	}

	go f.logWriter()
	go f.fileMonitor()

	fileLog = f
	return
}

// Log file are split.
func (f FileLogger) isMustSplit() bool {
	t, _ := time.Parse(DateFormat, time.Now().Format(DateFormat))
	return t.After(*f.date)
}

// Check if log files are exists, then create it.
func (f FileLogger) isExistOrCreate() {
	_, err := os.Stat(f.fileDir)
	if err != nil && !os.IsExist(err) {
		mkdirErr := os.Mkdir(f.fileDir, 0755)
		if mkdirErr != nil {
			log.Println("Create dir failed, error: ", mkdirErr)
		}
	}
}

// Split logs.
func (f *FileLogger) split() (err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	logFile := filepath.Join(f.fileDir, f.fileName)
	logFileBak := logFile + "." + f.date.Format(DateFormat)

	if f.logFile != nil {
		_ = f.logFile.Close()
	}

	err = os.Rename(logFile, logFileBak)
	if err != nil {
		return
	}

	t, _ := time.Parse(DateFormat, time.Now().Format(DateFormat))
	f.date = &t

	f.logFile, err = os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	f.lg = log.New(f.logFile, f.prefix, log.LstdFlags|log.Lmicroseconds)
	return
}

// Write to logs.
func (f *FileLogger) logWriter() {
	defer func() { recover() }()

	for {
		str := <-f.logChan

		f.mu.RLock()
		_ = f.lg.Output(2, str)
		f.mu.RUnlock()
	}
}

// Monitoring and split logs.
func (f *FileLogger) fileMonitor() {
	defer func() { recover() }()

	timer := time.NewTicker(30 * time.Second)
	for {
		<-timer.C

		if f.isMustSplit() {
			if err := f.split(); err != nil {
				f.Error("Log split error: %v\n", err)
			}
		}
	}
}

// Close logs.
func CloseLogger() {
	if fileLog != nil {
		close(fileLog.logChan)
		fileLog.lg = nil
		_ = fileLog.logFile.Close()
	}
}

// Output format logs.
func (f *FileLogger) Printf(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fileLog.logChan <- fmt.Sprintf("[%v:%v]", fmt.Sprintf(format, v...)+filepath.Base(file), line)
}

// Output format logs.
func (f *FileLogger) Print(v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fileLog.logChan <- fmt.Sprintf("[%v:%v]", fmt.Sprint(v...)+filepath.Base(file), line)
}

// Output format logs.
func (f *FileLogger) Println(v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fileLog.logChan <- fmt.Sprintf("[%v:%v]", filepath.Base(file), line) + fmt.Sprintln(v...)
}

// Output trace level logs.
func (f *FileLogger) Trace(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	if fileLog.logLevel <= TRACE {
		fileLog.logChan <- fmt.Sprintf("%v:%v]", fmt.Sprintf("[TRACE] [")+filepath.Base(file), line) + fmt.Sprintf(" "+format, v...)
	}
}

// Output info level logs.
func (f *FileLogger) Info(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	if fileLog.logLevel <= INFO {
		fileLog.logChan <- fmt.Sprintf("%v:%v]", fmt.Sprintf("[INFO] [")+filepath.Base(file), line) + fmt.Sprintf(" "+format, v...)
	}
}

// Output warn level logs.
func (f *FileLogger) Warn(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	if fileLog.logLevel <= WARN {
		fileLog.logChan <- fmt.Sprintf("%v:%v]", fmt.Sprintf("[WARN] [")+filepath.Base(file), line) + fmt.Sprintf(" "+format, v...)
	}
}

// Output error level logs.
func (f *FileLogger) Error(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	if fileLog.logLevel <= ERROR {
		fileLog.logChan <- fmt.Sprintf("%v:%v]", fmt.Sprintf("[ERROR] [")+filepath.Base(file), line) + fmt.Sprintf(" "+format, v...)
	}
}
