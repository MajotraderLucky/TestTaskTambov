package models

import (
	"fmt"

	"gopkg.in/reform.v1"
)

type NewsTableType struct{}

func (NewsTableType) Schema() string {
	return ""
}

// Изменение здесь
func (NewsTableType) Name() string {
	return "News"
}

func (NewsTableType) Columns() []string {
	return []string{"id", "title", "content"}
}

func (NewsTableType) NewStruct() reform.Struct {
	return new(NewsData)
}

var NewsTable = NewsTableType{}

// ...

type NewsData struct {
	ID         int64    `reform:"id"`
	Title      string   `reform:"title"`
	Content    string   `reform:"content"`
	Categories []string // нужно заранее позаботиться о заполнении этого поля
}

func (nd *NewsData) String() string {
	return fmt.Sprintf("NewsData<ID=%d, Title=%s, Content=%s>",
		nd.ID, nd.Title, nd.Content)
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

func fillNewsCategories(newsData *NewsData, categories []string) {
	newsData.Categories = categories
}

type NewsCategoryTableType struct{}

func (NewsCategoryTableType) Schema() string {
	return ""
}

// И изменение здесь
func (NewsCategoryTableType) Name() string {
	return "News_categories"
}

func (NewsCategoryTableType) Columns() []string {
	return []string{"news_id", "category_id"}
}

func (NewsCategoryTableType) NewStruct() reform.Struct {
	return new(NewsCategory)
}

var NewsCategoryTable = NewsCategoryTableType{}

// ...

type NewsCategory struct {
	NewsID     int64 `reform:"news_id"`
	CategoryID int64 `reform:"category_id"`
}

func (nc *NewsCategory) View() reform.View {
	return NewsCategoryTable
}

func (nc *NewsCategory) PKValues() []interface{} {
	return []interface{}{nc.NewsID, nc.CategoryID}
}

func (nc *NewsCategory) PKPointers() []interface{} {
	return []interface{}{&nc.NewsID, &nc.CategoryID}
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
