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

package mongo

import (
	"time"

	mgo "gopkg.in/mgo.v2"
)

var (
	ErrNotFound = mgo.ErrNotFound
)

type Interface interface {
	// Open 启动Mongo Driver
	Open(conf *Config) error

	// Close 关闭MongoDriver
	Close() error

	// GetSession 获取一个session, 这里的GetSession采用的是Copy的方式
	GetSession() *mgo.Session

	// PutSession 释放一个session
	PutSession(s *mgo.Session)
}

func New() Interface {
	return newMongoV1()
}

type Config struct {
	mgo.DialInfo
}

func DefaultConfig(addrs []string) *Config {
	return &Config{
		mgo.DialInfo{
			Addrs:    addrs,
			Timeout:  10 * time.Second,
			FailFast: true,
		},
	}
}

type mongoV1 struct {
	conf        *Config
	rootSession *mgo.Session
}

func newMongoV1() Interface {
	return &mongoV1{}
}

func (m *mongoV1) Open(conf *Config) (err error) {
	m.rootSession, err = mgo.DialWithInfo(&conf.DialInfo)
	if err != nil {
		return err
	}
	if err = m.rootSession.Ping(); err != nil {
		return err
	}
	return err
}

func (m *mongoV1) Close() error {
	if m.rootSession != nil {
		m.rootSession.Close()
	}
	return nil
}

func (m *mongoV1) GetSession() *mgo.Session {
	return m.rootSession.Copy()
}

func (m *mongoV1) PutSession(s *mgo.Session) {
	if s != nil {
		s.Close()
	}
}
