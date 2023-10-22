package newsdb

import (
	"testtasktambov/models"

	"gopkg.in/reform.v1"
)

// Создание новости и связывание ее с категориями
func CreateNewsWithCategories(session *reform.DB, newsData models.NewsData) error {
	tx, err := session.Begin()
	if err != nil {
		return err
	}

	// Создать новую новость
	news := &models.NewsData{
		Title:   newsData.Title,
		Content: newsData.Content,
	}

	if err := tx.Insert(news); err != nil {
		tx.Rollback()
		return err
	}

	for _, intCat := range newsData.Categories {
		// создание связей с категорией
		newsCategory := &models.NewsCategory{
			NewsID:     int64(news.ID),
			CategoryID: intCat,
		}

		if err := tx.Insert(newsCategory); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Зафиксировать транзакцию при успешном создании новости и связей
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
