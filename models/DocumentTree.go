package models

import (
	"bytes"
	"html/template"
	"math"
	"github.com/astaxie/beego/orm"
	"github.com/changming1987117/mindoc/conf"
	"fmt"
)

type DocumentTree struct {
	DocumentId   int               `json:"id"`
	DocumentName string            `json:"text"`
	ParentId     interface{}       `json:"parent"`
	Identify     string            `json:"identify"`
	BookIdentify string            `json:"-"`
	Version      int64             `json:"version"`
	State        *DocumentSelected `json:"-"`
	AAttrs		 map[string]interface{}		`json:"a_attr"`
}
type DocumentSelected struct {
	Selected bool `json:"selected"`
	Opened   bool `json:"opened"`
}

//获取项目的文档树状结构
func (item *Document) FindDocumentTree(bookId int, MemberId int) ([]*DocumentTree, error) {
	o := orm.NewOrm()

	trees := make([]*DocumentTree, 0)

	var docs []*Document

	count, err := o.QueryTable(item).Filter("book_id", bookId).OrderBy("order_sort", "document_id").Limit(math.MaxInt32).All(&docs, "document_id", "version", "document_name", "parent_id", "identify","is_open")

	if err != nil {
		return trees, err
	}
	book, _ := NewBook().Find(bookId)

	trees = make([]*DocumentTree, count)

	for index, item := range docs {
		var documentrelationship DocumentRelationship
		roleId, err := documentrelationship.FindForRoleId(item.DocumentId, MemberId)
		if err != nil && roleId == 4{
			continue
		}
		tree := &DocumentTree{}
		if index == 0 {
			tree.State = &DocumentSelected{Selected: true, Opened: true}
			tree.AAttrs = map[string]interface{}{ "is_open": true}
		}else if item.IsOpen == 1 {
			tree.State = &DocumentSelected{Selected: false, Opened: true}
			tree.AAttrs = map[string]interface{}{ "is_open": true}
		}
		tree.DocumentId = item.DocumentId
		tree.Identify = item.Identify
		tree.Version = item.Version
		tree.BookIdentify = book.Identify
		if item.ParentId > 0 {
			tree.ParentId = item.ParentId
		} else {
			tree.ParentId = "#"
		}

		tree.DocumentName = item.DocumentName

		trees[index] = tree
	}

	return trees, nil
}

func (item *Document) CreateDocumentTreeForHtml(bookId, selectedId int, MemberID int) (string, error) {
	trees, err := item.FindDocumentTree(bookId, MemberID)
	if err != nil {
		return "", err
	}
	parentId := getSelectedNode(trees, selectedId)

	buf := bytes.NewBufferString("")

	getDocumentTree(trees, 0, selectedId, parentId, buf)

	return buf.String(), nil

}

//使用递归的方式获取指定ID的顶级ID
func getSelectedNode(array []*DocumentTree, parent_id int) int {

	for _, item := range array {
		if _, ok := item.ParentId.(string); ok && item.DocumentId == parent_id {
			return item.DocumentId
		} else if pid, ok := item.ParentId.(int); ok && item.DocumentId == parent_id {
			return getSelectedNode(array, pid)
		}
	}
	return 0
}

func getDocumentTree(array []*DocumentTree, parentId int, selectedId int, selectedParentId int, buf *bytes.Buffer) {
	buf.WriteString("<ul>")

	for _, item := range array {
		pid := 0

		if p, ok := item.ParentId.(int); ok {
			pid = p
		}
		if pid == parentId {

			selected := ""
			if item.DocumentId == selectedId {
				selected = ` class="jstree-clicked"`
			}
			selectedLi := ""
			if item.DocumentId == selectedParentId || (item.State != nil && item.State.Opened) {
				selectedLi = ` class="jstree-open"`
			}
			buf.WriteString(fmt.Sprintf("<li id=\"%d\"%s><a href=\"",item.DocumentId,selectedLi))
			if item.Identify != "" {
				uri := conf.URLFor("DocumentController.Read", ":key", item.BookIdentify, ":id", item.Identify)
				buf.WriteString(uri)
			} else {
				uri := conf.URLFor("DocumentController.Read", ":key", item.BookIdentify, ":id", item.DocumentId)
				buf.WriteString(uri)
			}
			buf.WriteString(fmt.Sprintf("\" title=\"%s\"",template.HTMLEscapeString(item.DocumentName)))
			buf.WriteString(fmt.Sprintf(" data-version=\"%d\"%s>%s</a>",item.Version,selected,template.HTMLEscapeString(item.DocumentName)))


			for _, sub := range array {
				if p, ok := sub.ParentId.(int); ok && p == item.DocumentId {
					getDocumentTree(array, p, selectedId, selectedParentId, buf)
					break
				}
			}
			buf.WriteString("</li>")

		}
	}
	buf.WriteString("</ul>")
}
