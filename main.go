package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/kovalyov-valentin/go-fiber-postgres/models"
	"github.com/kovalyov-valentin/go-fiber-postgres/storage"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(ctx *fiber.Ctx) error {
	book := Book{}

	err := ctx.BodyParser(&book)

	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create book"})
		return err
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been added"})
	return nil
}

func (r *Repository) GetBooks(ctx *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could mot get books"})
		return err
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "books fetched successfully",
		"data":    bookModels,
	})
	return nil
}

func (r *Repository) DeleteBook(ctx *fiber.Ctx) error {
	bookModel := models.Books{}
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message":"id cannot be empty",
		})
		return nil
	}

	err := r.DB.Delete(bookModel, id)

	if err.Error != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message":"could not delete book",
		})
		return err.Error
	}
	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"books delete successfully",
	})
	return nil 
}

func (r *Repository) GetBooksById(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	bookModel := &models.Books{}

	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message":"id cannot be empty",
		})
		return nil 
	}

	fmt.Println("the Id is", id)

	err := r.DB.Where("id = ?", id).Find(bookModel).Error
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message":"could not get the book",
		})
		return err
	}
	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "books id fetched successfully", 
		"data": bookModel,
	})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBooksById)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("could not load the database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		DB: db,
	}

	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")

}
