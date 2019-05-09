package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/changming1987117/mindoc/conf"
)

type DocumentRelationship struct {
	RelationshipId int `orm:"pk;auto;unique;column(relationship_id)" json:"relationship_id"`
	MemberId       int `orm:"column(member_id);type(int)" json:"member_id"`
	DocumentId         int `orm:"column(document_id);type(int)" json:"document_id"`
	// RoleId 角色：0 创始人(创始人不能被移除) / 1 管理员/2 编辑者/3 观察者/4 游客
	RoleId conf.DocuementRole `orm:"column(role_id);type(int)" json:"role_id"`
}

// TableName 获取对应数据库表名.
func (m *DocumentRelationship) TableName() string {
	return "documentrelationship"
}
func (m *DocumentRelationship) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

// TableEngine 获取数据使用的引擎.
func (m *DocumentRelationship) TableEngine() string {
	return "INNODB"
}

// 联合唯一键
func (m *DocumentRelationship) TableUnique() [][]string {
	return [][]string{
		{"member_id", "document_id"},
	}
}

func (m *DocumentRelationship) QueryTable() orm.QuerySeter  {
	return orm.NewOrm().QueryTable(m.TableNameWithPrefix())
}
func NewDocumentRelationship() *DocumentRelationship {
	return &DocumentRelationship{}
}

func (m *DocumentRelationship) Find(id int) (*DocumentRelationship, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("relationship_id", id).One(m)
	return m, err
}

func (m *DocumentRelationship) FindForRoleId(documentId int, memberId int) (conf.DocuementRole, error) {
	o := orm.NewOrm()

	relationship := NewDocumentRelationship()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("document_id", documentId).Filter("member_id", memberId).One(relationship)

	if err != nil {

		return 0, err
	}
	return relationship.RoleId, nil
}

func (m *DocumentRelationship) FindByBookIdAndMemberId(bookId, memberId int) (*DocumentRelationship, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("document_id", bookId).Filter("member_id", memberId).One(m)

	return m, err
}

func (m *DocumentRelationship) Insert() error {
	o := orm.NewOrm()

	_, err := o.Insert(m)

	return err
}

func (m *DocumentRelationship) Update() error {
	o := orm.NewOrm()

	_, err := o.Update(m)

	return err
}

