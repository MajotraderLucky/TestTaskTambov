package models

import (
	"fmt"
	"log"

	"gopkg.in/reform.v1"
)

type NewsTableType struct {
	s *reform.Struct
}

func (nt NewsTableType) PKColumnIndex() uint {
	return 0 // индекс столбца, который является первичным ключом
}

func (NewsTableType) Schema() string {
	return "" // or return your actual schema if you have one
}

func (NewsTableType) Name() string {
	return "News"
}

func (NewsTableType) Columns() []string {
	return []string{"id", "title", "content"}
}

func (NewsTableType) NewStruct() reform.Struct {
	return new(NewsData)
}

func (NewsTableType) NewRecord() reform.Record {
	return new(NewsData)
}

var NewsTable = NewsTableType{}

type NewsCategoryTableType struct {
	pkColumnIndex uint
}

func (nct NewsCategoryTableType) PKColumnIndex() uint {
	return nct.pkColumnIndex
}

func (NewsCategoryTableType) NewRecord() reform.Record {
	return new(NewsCategory)
}

func (NewsCategoryTableType) Schema() string {
	return "" // or return your actual schema if you have one
}

func (NewsCategoryTableType) Name() string {
	return "news_categories"
}

func (NewsCategoryTableType) Columns() []string {
	return []string{"news_id", "category_id"}
}

func (NewsCategoryTableType) NewStruct() reform.Struct {
	return new(NewsCategory)
}

var NewsCategoryTable = NewsCategoryTableType{}

var Table_name = "News"

type NewsData struct {
	ID         int64  `reform:"id,pk"`
	Title      string `reform:"title"`
	Content    string `reform:"content"`
	Categories []int64
}

func (nd *NewsData) View() reform.View {
	return NewsTable
}

func (nd *NewsData) PKValue() interface{} {
	return nd.ID
}

func (nd *NewsData) PKPointer() interface{} {
	return &nd.ID
}

func (nd *NewsData) HasPK() bool {
	return nd.ID != 0
}

func (nd *NewsData) Values() []interface{} {
	return []interface{}{
		nd.ID,
		nd.Title,
		nd.Content,
	}
}

func (nd *NewsData) Pointers() []interface{} {
	return []interface{}{
		&nd.ID,
		&nd.Title,
		&nd.Content,
	}
}

func (nd *NewsData) String() string {
	return fmt.Sprintf("NewsData {ID: %d, Title: %s, Content: %s}", nd.ID, nd.Title, nd.Content)
}

func (nd *NewsData) SetPK(pk interface{}) {
	pkInt, ok := pk.(int64)
	if !ok {
		log.Panicln("SetPK: provided value is not int64")
	}
	nd.ID = pkInt
}

func (nd *NewsData) Table() reform.Table {
	return NewsTable
}

type NewsCategory struct {
	NewsID     int64 `reform:"news_id"`
	CategoryID int64 `reform:"category_id"`
}

func (nc *NewsCategory) Table() reform.Table {
	return NewsCategoryTable
}

func (nc *NewsCategory) View() reform.View {
	return NewsCategoryTable
}

func (nc *NewsCategory) PKValue() interface{} {
	return []interface{}{nc.NewsID, nc.CategoryID}
}

func (nc *NewsCategory) PKPointer() interface{} {
	return []interface{}{&nc.NewsID, &nc.CategoryID}
}

func (nc *NewsCategory) HasPK() bool {
	return nc.NewsID != 0 && nc.CategoryID != 0
}

func (nc *NewsCategory) Values() []interface{} {
	return []interface{}{
		nc.NewsID,
		nc.CategoryID,
	}
}

func (nc *NewsCategory) Pointers() []interface{} {
	return []interface{}{
		&nc.NewsID,
		&nc.CategoryID,
	}
}

func (nc *NewsCategory) String() string {
	return fmt.Sprintf("NewsCategory<NewsID=%d, CategoryID=%d>", nc.NewsID, nc.CategoryID)
}

func (nc *NewsCategory) SetPK(pk interface{}) {
	// Задаем предположение, что входящий ключ - это
	// массив или срез из двух элементов типа int64.
	keys, ok := pk.([]int64)
	if ok && len(keys) == 2 {
		nc.NewsID = keys[0]
		nc.CategoryID = keys[1]
	} else {
		panic("SetPK: unsupported format")
	}
}
