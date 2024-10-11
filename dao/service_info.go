package dao

import (
	"fmt"
	"gatewat_web/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type ServiceInfo struct {
	ID          int64     `json:"id" gorm:"primary_key" description:"自增主键" `
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载均衡 0=http 1=tcp 2=grpc" `
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称" `
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述" `
	UpdatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间" `
	CreatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间" `
	IsDelete    int8      `json:"is_delete" gorm:"column:is_delete" description:"是否删除" `
}

// 指定表名
func (serviceInfo *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (serviceInfo *ServiceInfo) GetServiceDetailByServiceInfo(c *gin.Context, db *gorm.DB, search *ServiceInfo) (*ServiceDetail, error) {
	if search.ServiceName == "" {
		info, err := serviceInfo.Find(c, db, search)
		if err != nil {
			return nil, err
		}
		search = info
	}
	httpRule := &HttpRule{ServiceID: search.ID}
	httpRule, err := httpRule.Find(c, db, httpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	tcpRule := &TcpRule{ServiceID: search.ID}
	tcpRule, err = tcpRule.Find(c, db, tcpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	grpcRule := &GrpcRule{ServiceID: search.ID}
	grpcRule, err = grpcRule.Find(c, db, grpcRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	accessControl := &AccessControl{ServiceID: search.ID}
	accessControl, err = accessControl.Find(c, db, accessControl)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	loadBalance := &LoadBalance{ServiceID: search.ID}
	loadBalance, err = loadBalance.Find(c, db, loadBalance)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	detail := &ServiceDetail{
		Info:          search,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
	}
	return detail, nil
}

func (serviceInfo *ServiceInfo) PageList(c *gin.Context, db *gorm.DB, param *dto.ServiceListInput) ([]ServiceInfo, int64, error) {
	total := int64(0)
	list := []ServiceInfo{}
	//偏移量是查询的起始位置
	offset := (param.PageNum - 1) * param.PageSize
	//下面的查询语句等价于sql语句：
	//SELECT * FROM service_info
	//WHERE is_delete = 0
	//  AND (service_name LIKE '%关键字%' OR service_desc LIKE '%关键字%')
	//ORDER BY id DESC
	//LIMIT 每页大小 OFFSET 偏移量;
	query := db.WithContext(c)
	query = query.Table(serviceInfo.TableName()).Where("is_delete=0")
	if param.KeyWord != "" {
		query = query.Where("(service_name like ? or service_desc like ?)", "%"+param.KeyWord+"%", "%"+param.KeyWord+"%")
	}
	if err := query.Limit(param.PageSize).Offset(offset).Order("id desc").Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	query.Limit(param.PageSize).Offset(offset).Count(&total)
	return list, total, nil
}

func (serviceInfo *ServiceInfo) Find(c *gin.Context, db *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	out := &ServiceInfo{}
	err := db.WithContext(c).Where(search).First(out).Error
	fmt.Println(out)
	if err != nil {
		fmt.Println("err = ", err)
		return nil, err
	}
	return out, nil
}

func (serviceInfo *ServiceInfo) Save(c *gin.Context, db *gorm.DB) error {
	err := db.WithContext(c).Save(serviceInfo).Error
	if err != nil {
		fmt.Println("err = ", err)
		return err
	}
	return nil
}

func (serviceInfo *ServiceInfo) GroupByLoadType(c *gin.Context, db *gorm.DB) ([]dto.DashBoardServiceStatItemOutput, error) {
	list := []dto.DashBoardServiceStatItemOutput{}
	query := db.WithContext(c)
	if err := query.Table(serviceInfo.TableName()).Where("is_delete=0").Select("load_type, count(*) as value").Group("load_type").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
