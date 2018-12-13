// Copyright 2018 Hurricanezwf. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Hurricanezwf/pkg/logging/glog"
)

const (
	LogWayConsole = "console"
	LogWayFile    = "file"
)

var (
	mutex   sync.RWMutex
	logWay  string        = LogWayConsole
	logDir  string        = ""
	verbose int           = 1
	expire  time.Duration = time.Duration(7 * 24 * time.Hour)
)

func init() {
	reset(logWay, logDir, verbose)
	go cleanDaemon()
}

func Reset(logway, logdir string, verboselevel int) (err error) {
	if err = reset(logway, logdir, verboselevel); err != nil {
		return err
	}
	flag.Parse() // do apply to glog
	return nil
}

func reset(logway, logdir string, verboselevel int) (err error) {
	if verboselevel < 0 {
		return errors.New("VerboseLevel must be >= 0")
	}

	switch logway {
	case LogWayConsole:
		flag.Set("logtostderr", "true")
	case LogWayFile:
		if len(logdir) <= 0 {
			return errors.New("Missing logDir when you want to log into file")
		}
		if logdir, err = filepath.Abs(logdir); err != nil {
			return fmt.Errorf("Make abs logdir failed, %v", err)
		}
		if err = os.MkdirAll(logdir, os.ModePerm); err != nil {
			return fmt.Errorf("Create log dir failed, %v", err)
		}
		flag.Set("logtostderr", "false")
		flag.Set("log_dir", logdir)
	default:
		return fmt.Errorf("Unknown logway(%s)", logway)
	}

	mutex.Lock()
	logWay = logway
	logDir = logdir
	verbose = verbose
	mutex.Unlock()

	flag.Set("stderrthreshold", "ERROR")
	flag.Set("v", strconv.Itoa(verboselevel))
	return nil
}

func LogWayOK(logway string) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return (logway == LogWayConsole || logway == LogWayFile)
}

func SetExpire(v time.Duration) error {
	if v < 10*time.Minute {
		return errors.New("logging: Expire duration should be greater than 10 minutes")
	}
	mutex.Lock()
	expire = v
	mutex.Unlock()
	return nil
}

func Expire() time.Duration {
	mutex.RLock()
	defer mutex.RUnlock()
	return expire
}

func LogWay() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return logWay
}

func LogDir() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return logDir
}

// 定期清理日志
func cleanDaemon() {
	batchLimit := 5
	toRemove := make([]string, 0, batchLimit)

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if LogWay() != LogWayFile {
			continue
		}

		// 筛选出需要清理的日志, 每次清理总数目受到batchLimit限制
		// 可清理的日志需满足以下条件:
		// (1) 过期
		// (2) 具备文件名中具备.log的关键字
		// (3) 具备写权限
		toRemove = toRemove[:0]
		keepDuration := Expire()
		dir := LogDir()

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			glog.Warning(err.Error())
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			modeTime := f.ModTime()
			if time.Since(modeTime) < keepDuration {
				continue
			}
			if strings.Contains(f.Name(), ".log") == false {
				continue
			}
			if (f.Mode() & 0200) == 0 {
				continue
			}
			toRemove = append(toRemove, filepath.Join(dir, f.Name()))
			if len(toRemove) >= batchLimit {
				break
			}
		}

		for _, f := range toRemove {
			if err := os.Remove(f); err != nil {
				glog.Warningf("Clean %s failed, %v", f, err)
				continue
			}
		}
	}
}

// DebugWriter 输出Debug级别日志
type DebugWriter struct{}

func (w DebugWriter) Println(format string, v ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintf(format, v...))
}

// InfoWriter 输出Info级别日志
type InfoWriter struct{}

func (w InfoWriter) Println(format string, v ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintf(format, v...))
}

// WarnWriter 输出Warn级别日志
type WarnWriter struct{}

func (w WarnWriter) Println(format string, v ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintf(format, v...))
}

// ErrorWriter 输出Error级别日志
type ErrorWriter struct{}

func (w ErrorWriter) Println(format string, v ...interface{}) {
	glog.ErrorDepth(1, fmt.Sprintf(format, v...))
}

// 如需更详细的级别，可直接使用glog.V(n).Infof()
