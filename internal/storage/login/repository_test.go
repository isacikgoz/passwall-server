package login_test

import (
	"database/sql"
	"database/sql/driver"
	"regexp"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pass-wall/passwall-server/internal/storage/login"
	. "github.com/pass-wall/passwall-server/model"
)

var _ = Describe("Login", func() {
	var repository *login.Repository
	var mock sqlmock.Sqlmock

	BeforeEach(func() {
		var db *sql.DB
		var err error

		// db, mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual)) // use equal matcher
		db, mock, err = sqlmock.New() // mock sql.DB
		Expect(err).ShouldNot(HaveOccurred())

		gdb, err := gorm.Open("postgres", db) // open gorm db
		Expect(err).ShouldNot(HaveOccurred())

		repository = login.NewRepository(gdb)
	})
	AfterEach(func() {
		err := mock.ExpectationsWereMet() // make sure all expectations were met
		Expect(err).ShouldNot(HaveOccurred())
	})

	// FIND ALL TEST
	Context("find all", func() {
		It("empty", func() {
			const sqlSelectAll = `SELECT * FROM "logins"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelectAll)).
				WillReturnRows(sqlmock.NewRows(nil))
			argsStr := map[string]string{
				"search": "",
				"order":  "",
			}

			argsInt := map[string]int{
				"offset": -1,
				"limit":  -1,
			}
			l, err := repository.FindAll(argsStr, argsInt)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(l).Should(BeEmpty())
		})
	})

	Context("find by id", func() {

		// FOUND BY ID TEST
		It("found", func() {
			login := Login{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: nil,
				URL:       "https://dummywebsite.com",
				Username:  "DummyUser",
				Password:  "DummyPassword",
			}

			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "url", "username", "password"}).
				AddRow(login.ID, login.CreatedAt, login.UpdatedAt, login.DeletedAt, login.URL, login.Username, login.Password)

			const sqlSelectOne = `SELECT * FROM "logins" WHERE "logins"."deleted_at" IS NULL AND ((id = $1)) ORDER BY "logins"."id" ASC LIMIT 1`

			mock.ExpectQuery(regexp.QuoteMeta(sqlSelectOne)).
				WithArgs(login.ID).
				WillReturnRows(rows)

			dbLogin, err := repository.FindByID(login.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(dbLogin).Should(Equal(login))
		})

		// NOT FOUND BY ID TEST
		It("not found", func() {
			// ignore sql match
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			_, err := repository.FindByID(1)
			Expect(err).Should(Equal(gorm.ErrRecordNotFound))
		})
	})

	Context("save", func() {
		var login Login
		BeforeEach(func() {
			login = Login{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: nil,
				URL:       "https://dummywebsite.com",
				Username:  "DummyUser",
				Password:  "DummyPassword",
			}
		})

		// UPDATE TEST
		It("update", func() {
			const sqlUpdate = `UPDATE "logins" SET "created_at" = $1, "updated_at" = $2, "deleted_at" = $3, "url" = $4, "username" = $5, "password" = $6 WHERE "logins"."deleted_at" IS NULL AND "logins"."id" = $7`
			const sqlSelectOne = `SELECT * FROM "logins" WHERE "logins"."deleted_at" IS NULL AND "logins"."id" = $1 ORDER BY "logins"."id" ASC LIMIT 1`

			login.ID = 1
			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).
				WithArgs(AnyTime{}, AnyTime{}, login.DeletedAt, login.URL, login.Username, login.Password, login.ID).
				WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectCommit()

			// select after update
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelectOne)).
				WithArgs(login.ID).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(login.ID))

			_, err := repository.Save(login)
			Expect(err).ShouldNot(HaveOccurred())
		})

		// INSERT TEST
		It("insert", func() {
			// gorm use query instead of exec
			// https://github.com/DATA-DOG/go-sqlmock/issues/118
			const sqlInsert = `
					INSERT INTO "logins" ("created_at","updated_at","deleted_at","url","username","password")
						VALUES ($1,$2,$3,$4,$5,$6) RETURNING "logins"."id"`
			const newId = 1
			mock.ExpectBegin() // start transaction
			mock.ExpectQuery(regexp.QuoteMeta(sqlInsert)).
				WithArgs(login.CreatedAt, login.UpdatedAt, login.DeletedAt, login.URL, login.Username, login.Password).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(newId))
			mock.ExpectCommit() // commit transaction

			Expect(login.ID).Should(BeZero())

			savedLogin, err := repository.Save(login)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(savedLogin.ID).Should(BeEquivalentTo(newId))
		})

	})

})

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}
