package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/n-korel/social-api/internal/store"
)

var usernames = []string{
	"Nikita", "Ksenia", "Aleksandr", "Maria", "Dmitrii", "Anna", "Sergei", "Olga",
	"Ivan", "Ekaterina", "Andrei", "Tatiana", "Mikhail", "Natalia", "Vladimir", "Elena",
	"Pavel", "Svetlana", "Aleksei", "Irina", "Artem", "Iulia", "Maksim", "Galina",
	"Roman", "Viktoria", "Igor", "Liudmila", "Vitalii", "Nadezhda", "Evgenii", "Valentina",
	"Konstantin", "Margarita", "Oleg", "Anastasia", "Vladislav", "Sofia", "Denis", "Alla",
	"Nikolai", "Vera", "Fedor", "Lidiia", "Georgii", "Zinaida", "Boris", "Raisa",
	"Vasilii", "Tamara",
}

var titles = []string{
	"Go за 5 минут", "Микросервисы на Go", "Concurrency patterns", "Go vs Rust",
	"REST API с Gin", "Docker и Go", "Тестирование в Go", "Go modules guide",
	"Производительность Go", "GraphQL на Go", "Go и PostgreSQL",
	"Channels и goroutines", "Clean Architecture", "Go для начинающих",
	"Debugging в Go", "Go и Kubernetes", "Middleware patterns",
	"JWT авторизация", "Go best practices", "Профилирование Go",
}

var contents = []string{
	"Краткое введение в язык Go: синтаксис, компиляция, основные типы данных и структуры управления. Идеальный старт для новичков.",
	"Как построить масштабируемую архитектуру микросервисов используя Go. Паттерны, инструменты и best practices.",
	"Основные паттерны параллельного программирования в Go: worker pools, fan-in/fan-out, pipeline, context cancellation.",
	"Детальное сравнение Go и Rust: производительность, безопасность памяти, экосистема. Когда выбрать каждый язык.",
	"Создание RESTful API с использованием Gin framework. Роутинг, middleware, валидация данных и обработка ошибок.",
	"Контейнеризация Go-приложений: multi-stage builds, оптимизация размера образа, best practices для production.",
	"Полное руководство по тестированию: unit-тесты, интеграционные тесты, mocking, coverage и бенчмарки.",
	"Управление зависимостями с Go modules: создание модулей, версионирование, replace директивы, vendoring.",
	"Оптимизация производительности Go-приложений: memory allocation, garbage collector, CPU profiling.",
	"Создание GraphQL API на Go с использованием gqlgen. Схемы, резолверы, subscriptions и интеграция с БД.",
	"Работа с PostgreSQL в Go: подключение, миграции, CRUD операции, connection pooling, transactions.",
	"Глубокое погружение в concurrency модель Go: создание goroutines, типы каналов, select statement.",
	"Реализация Clean Architecture в Go проектах: разделение слоев, dependency injection, тестируемый код.",
	"Пошаговое руководство для новичков: установка, первая программа, основы языка, полезные ресурсы.",
	"Инструменты и техники отладки Go-приложений: Delve debugger, логирование, race detector, pprof.",
	"Разработка cloud-native приложений на Go для Kubernetes: health checks, graceful shutdown, операторы.",
	"Паттерны middleware в Go веб-приложениях: логирование, аутентификация, rate limiting, CORS.",
	"Реализация JWT-авторизации в Go: создание токенов, валидация, refresh tokens, безопасность.",
	"Лучшие практики разработки на Go: структура проекта, именование, обработка ошибок, code style.",
	"Анализ производительности Go-приложений: CPU и memory profiling, flame graphs, оптимизация bottlenecks.",
}

var tags = []string{
	"golang", "programming", "backend", "api", "microservices",
	"concurrency", "performance", "testing", "docker", "kubernetes",
	"database", "postgresql", "rest", "graphql", "gin",
	"middleware", "jwt", "security", "architecture", "patterns",
}

var comments = []string{
	"Отличная статья! Очень помогла разобраться с goroutines",
	"Спасибо за подробные примеры кода",
	"А можно добавить пример с context.WithTimeout?",
	"Понятно объяснили сложные концепции",
	"Docker multi-stage builds действительно экономят место",
	"Наконец-то понял как работают каналы",
	"Gin гораздо проще чем стандартный http пакет",
	"Хороший гайд по тестированию, но не хватает примеров с mock",
	"pprof - мощный инструмент, спасибо за введение",
	"Kubernetes и Go - идеальная связка для микросервисов",
	"PostgreSQL connection pooling очень важная тема",
	"JWT реализация выглядит безопасно",
	"Clean Architecture в Go действительно работает",
	"Отлично подходит для новичков",
	"Delve debugger изменил мою жизнь",
	"Middleware паттерны очень полезные",
	"gqlgen - отличный выбор для GraphQL",
	"Go modules наконец-то решили проблему зависимостей",
	"Профилирование - must have для production",
	"Эти практики действительно работают в реальных проектах",
}

func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()

	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			log.Println("Error creating user:", err)
			return
		}
	}

	tx.Commit()

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating post:", err)
			return
		}
	}

	comments := generateComments(500, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("Error creating comment:", err)
			return
		}
	}

	log.Println("Seeding complete")
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)

	for i := 0; i < num; i++ {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Role: store.Role{
				Name: "user",
			},
		}
	}

	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)
	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]

		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: contents[rand.Intn(len(contents))],
			Tags: []string{
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
			},
		}

	}

	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	cms := make([]*store.Comment, num)
	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]
		post := posts[rand.Intn(len(posts))]

		cms[i] = &store.Comment{
			UserID:  user.ID,
			PostID:  post.ID,
			Content: comments[rand.Intn(len(comments))],
		}

	}

	return cms
}
