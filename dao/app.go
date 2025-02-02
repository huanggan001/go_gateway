package dao

import (
	"gatewat_web/common/lib"
	"gatewat_web/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http/httptest"
	"sync"
	"time"
)

type App struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (app *App) TableName() string {
	return "gateway_app"
}

func (app *App) Find(c *gin.Context, db *gorm.DB, search *App) (*App, error) {
	model := &App{}
	err := db.WithContext(c).Where(search).First(model).Error
	return model, err
}

func (app *App) Save(c *gin.Context, db *gorm.DB) error {
	err := db.WithContext(c).Save(app).Error
	return err
}

func (app *App) APPList(c *gin.Context, db *gorm.DB, params *dto.AppListInput) ([]App, int64, error) {
	var list []App
	var count int64
	pageNum := params.PageNum
	pageSize := params.PageSize

	//limit offset,pagesize
	offset := (pageNum - 1) * pageSize
	query := db.WithContext(c)
	query = query.Table(app.TableName()).Select("*")
	query = query.Where("is_delete=?", 0)
	if params.KeyWord != "" {
		query = query.Where("(name like ? or app_id like ?)", "%"+params.KeyWord+"%", "%"+params.KeyWord+"%")
	}
	err := query.Limit(pageSize).Offset(offset).Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	errCount := query.Count(&count).Error
	if errCount != nil {
		return nil, 0, errCount
	}
	return list, count, nil
}

var AppManagerHandler *AppManager

func init() {
	AppManagerHandler = NewAppManager()
}

type AppManager struct {
	AppMap   map[string]*App
	AppSlice []*App
	Locker   sync.RWMutex
	init     sync.Once
	err      error
}

func NewAppManager() *AppManager {
	return &AppManager{
		AppMap:   map[string]*App{},
		AppSlice: []*App{},
		Locker:   sync.RWMutex{},
		init:     sync.Once{},
	}
}

func (s *AppManager) GetAppList() []*App {
	return s.AppSlice
}

func (s *AppManager) LoadOnce() error {
	s.init.Do(func() {
		appInfo := &App{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}

		params := &dto.AppListInput{PageNum: 1, PageSize: 99999}
		list, _, err := appInfo.APPList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}

		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			s.AppMap[listItem.AppID] = &tmpItem
			s.AppSlice = append(s.AppSlice, &tmpItem)
		}
	})
	return s.err
}
