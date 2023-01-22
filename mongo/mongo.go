package mongo

import (
	"gitlab.dian.org.cn/helper/miniapp-platform/logger"
	mgo "gopkg.in/mgo.v2"
)

var _defaultDB *mgo.Session

func Init(host, port, dbname, user, password string) error {
	mgosession, err := mgo.Dial(host + ":" + port)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		logger.GetLogger().Panic("panic: connect wrong")
		return err
	}
	mgosession.SetMode(mgo.Monotonic, true)
	mgosession.SetPoolLimit(300)
	myDb := mgosession.DB(dbname)
	if user != "" {
		err = myDb.Login(user, password)
		if err != nil {
			logger.GetLogger().Error(err.Error())
			logger.GetLogger().Panic("panic: login db wrong")
			return err
		}
	}
	_defaultDB = mgosession
	return nil
}

func GetMongoDB() *mgo.Session {
	return _defaultDB
}
