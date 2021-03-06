package gosql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"
)

var (
	createSchema = `
CREATE TABLE users (
	id int(11) unsigned NOT NULL AUTO_INCREMENT,
	name  varchar(100) NOT NULL DEFAULT '',
	email  varchar(100) NOT NULL DEFAULT '',
	status  int(11) NOT NULL DEFAULT 0,
    success_time datetime DEFAULT NULL,
	created_at datetime NOT NULL,
	updated_at datetime NOT NULL,
  	PRIMARY KEY (id)
)ENGINE=InnoDB CHARSET=utf8;
`

	dropSchema = `
	drop table users
`

	createSchema2 = `
CREATE TABLE posts (
	id int(11) unsigned NOT NULL AUTO_INCREMENT,
	title  varchar(100) NOT NULL DEFAULT '',
	content  varchar(100) NOT NULL DEFAULT '',
	status  int(11) NOT NULL DEFAULT 0,
	created_at datetime NOT NULL,
	updated_at datetime NOT NULL,
  	PRIMARY KEY (id)
)ENGINE=InnoDB CHARSET=utf8;
`

	dropSchema2 = `
	drop table posts
`
)

type Users struct {
	Id          int            `db:"id"`
	Name        string         `db:"name"`
	Email       string         `db:"email"`
	Status      int            `db:"status"`
	SuccessTime sql.NullString `db:"success_time" json:"success_time"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

func (u *Users) DbName() string {
	return "default"
}

func (u *Users) TableName() string {
	return "users"
}

func (u *Users) PK() string {
	return "id"
}

type Posts struct {
	Id        int       `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	Status    int       `db:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (u *Posts) DbName() string {
	return "db2"
}

func (u *Posts) TableName() string {
	return "posts"
}

func (u *Posts) PK() string {
	return "id"
}

func RunWithSchema(t *testing.T, test func(t *testing.T)) {
	db := DB()
	defer func() {
		_, err := db.Exec(dropSchema)
		if err != nil {
			t.Error(err)
		}
	}()

	_, err := db.Exec(createSchema)

	if err != nil {
		t.Fatalf("create schema error:%s", err)
	}

	test(t)
}

func RunWithSchema2(t *testing.T, test func(t *testing.T)) {
	db := DB("db2")
	defer func() {
		_, err := db.Exec(dropSchema2)
		if err != nil {
			t.Error(err)
		}
	}()

	_, err := db.Exec(createSchema2)

	if err != nil {
		t.Fatalf("create schema error:%s", err)
	}

	test(t)
}

func insert(id int) {
	user := &Users{
		Id:     id,
		Name:   "test" + strconv.Itoa(id),
		Status: 1,
		Email:  "test" + strconv.Itoa(id) + "@test.com",
	}
	_, err := Model(user).Create()
	if err != nil {
		log.Fatal(err)
	}
}

func insertStatus(id int, status int) {
	user := &Users{
		Id:     id,
		Name:   "test" + strconv.Itoa(id),
		Status: status,
		Email:  "test" + strconv.Itoa(id) + "@test.com",
	}
	Model(user).Create()
}

func TestBuilder_Get(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		{
			user := &Users{}
			err := Model(user).Where("id = ?", 1).Get()

			if err != nil {
				t.Error(err)
			}
			//fmt.Println(user)

		}

		{
			user := &Users{
				Name:   "test1",
				Status: 1,
			}
			err := Model(user).Get()

			if err != nil {
				t.Error(err)
			}
			fmt.Println(user)
		}

		{
			insertStatus(2, 0)
			user := &Users{
				Status: 0,
			}

			err := Model(user).Where("id = ?", 2).Get("status")

			if err != nil {
				t.Error(err)
			}
			fmt.Println(user)
		}
	})
}

func TestBuilder_Hint(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		insert(2)

		user := make([]*Users, 0)
		err := Model(&user).Hint("/*+TDDL:slave()*/").All()

		if err != nil {
			t.Error(err)
		}

		fmt.Println(jsonEncode(user))
	})
}

func jsonEncode(i interface{}) string {
	ret, _ := json.Marshal(i)
	return string(ret)
}

func TestBuilder_All(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		insert(2)

		user := make([]*Users, 0)
		err := Model(&user).All()

		if err != nil {
			t.Error(err)
		}

		fmt.Println(jsonEncode(user))
	})
}

func TestBuilder_InAll(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		insert(2)
		insert(3)
		insert(4)
		insert(5)

		user := make([]*Users, 0)
		err := Model(&user).Where("status = ? and id in(?)",1,[]int{1,3,4}).All()

		if err != nil {
			t.Error(err)
		}

		fmt.Println(jsonEncode(user))
	})
}

func TestBuilder_Update(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)

		{
			user := &Users{
				Name: "test2",
			}

			affected, err := Model(user).Where("id=?", 1).Update()

			if err != nil {
				t.Error("update user error", err)
			}

			if affected == 0 {
				t.Error("update user affected error", err)
			}
		}

		{
			user := &Users{
				Id:   1,
				Name: "test3",
			}

			affected, err := Model(user).Update()

			if err != nil {
				t.Error("update user error", err)
			}

			if affected == 0 {
				t.Error("update user affected error", err)
			}
		}
	})
}

func TestBuilder_Delete(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		{
			insert(1)
			affected, err := Model(&Users{}).Where("id=?", 1).Delete()

			if err != nil {
				t.Error("delete user error", err)
			}

			if affected == 0 {
				t.Error("delete user affected error", err)
			}
		}
		{
			insert(1)
			affected, err := Model(&Users{Id: 1}).Delete()

			if err != nil {
				t.Error("delete user error", err)
			}

			if affected == 0 {
				t.Error("delete user affected error", err)
			}
		}

		{
			insertStatus(1, 0)
			insertStatus(2, 0)
			insertStatus(3, 0)

			affected, err := Model(&Users{Status: 0}).Delete("status")

			if err != nil {
				t.Error("delete user error", err)
			}

			if affected != 3 {
				t.Error("delete user affected error", err)
			}
		}
	})
}

func TestBuilder_Count(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		{
			num, err := Model(&Users{}).Count()

			if err != nil {
				t.Error(err)
			}

			if num != 1 {
				t.Error("count user error")
			}
		}

		{
			insertStatus(2, 0)
			insertStatus(3, 0)

			num, err := Model(&Users{Status: 0}).Count("status")

			if err != nil {
				t.Error(err)
			}

			if num != 2 {
				t.Error("count user error")
			}
		}
	})
}

func TestBuilder_Create(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		user := &Users{
			//Id:    1,
			Name:  "test",
			Email: "test@test.com",
		}
		id, err := Model(user).Create()

		if err != nil {
			t.Error(err)
		}

		if id != 1 {
			t.Error("lastInsertId error", id)
		}

		if int(id) != user.Id {
			t.Error("fill primaryKey error", id)
		}
	})
}

func TestBuilder_Limit(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		insert(2)
		insert(3)
		user := &Users{}
		err := Model(user).Limit(1).Get()

		if err != nil {
			t.Error(err)
		}
	})
}

func TestBuilder_Offset(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		insert(2)
		insert(3)
		user := &Users{}
		err := Model(user).Limit(1).Offset(1).Get()

		if err != nil {
			t.Error(err)
		}
	})
}

func TestBuilder_OrderBy(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		insert(2)
		insert(3)
		user := &Users{}
		err := Model(user).OrderBy("id desc").Limit(1).Offset(1).Get()

		if err != nil {
			t.Error(err)
		}

		if user.Id != 2 {
			t.Error("order by error")
		}

		//fmt.Println(user)
	})
}

func TestBuilder_Where(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		insert(1)
		insert(2)
		insert(3)
		user := make([]*Users, 0)
		err := Model(&user).Where("id in(?,?)", 2, 3).OrderBy("id desc").All()

		if err != nil {
			t.Error(err)
		}

		if len(user) != 2 {
			t.Error("where error")
		}

		//fmt.Println(user)
	})
}

type TimeFields struct {
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type UserCombs struct {
	Id          int            `db:"id"`
	Name        string         `db:"name"`
	Email       string         `db:"email"`
	Status      int            `db:"status"`
	SuccessTime sql.NullString `db:"success_time" json:"success_time"`
	TimeFields
}

func (u *UserCombs) DbName() string {
	return "default"
}

func (u *UserCombs) TableName() string {
	return "users"
}

func (u *UserCombs) PK() string {
	return "id"
}

func TestBuilder_NullString(t *testing.T) {
	RunWithSchema(t, func(t *testing.T) {
		ct, _ := time.Parse("2006-01-02 15:04:05", "2018-09-02 00:00:00")
		{
			user := &Users{
				Id:     1,
				Name:   "test",
				Status: 1,
				SuccessTime: sql.NullString{
					String: "2018-09-03 00:00:00",
					Valid:  true,
				},
				Email:     "test@test.com",
				CreatedAt: ct,
			}
			_, err := Model(user).Create()
			if err != nil {
				log.Fatal(err)
			}
		}

		{
			user := &Users{}
			err := Model(user).Where("id=1").Get()

			if err != nil {
				t.Error(err)
			}

			fmt.Println(jsonEncode(user))
		}

		{
			user := &Users{
				Id: 1,
				SuccessTime: sql.NullString{
					String: "2018-09-03 00:00:00",
					Valid:  true,
				},
				CreatedAt: ct,
			}

			err := Model(user).Get()

			if err != nil {
				t.Error(err)
			}

			fmt.Println(jsonEncode(user))
		}

		{
			user := &UserCombs{
				Id: 1,
				TimeFields: TimeFields{
					CreatedAt: ct,
				},
			}
			err := Model(user).Get()

			if err != nil {
				t.Error(err)
			}

			fmt.Println(jsonEncode(user))
		}
	})
}

func TestBuilder_Relation1(t *testing.T) {
	moment := &MomentList{}
	err := Model(moment).Relation("User" , func(b *Builder) {
		b.Where("gender = 1")
	}).Where("status = 1 and id = ?",14).Get()

	b , _ :=json.MarshalIndent(moment,"","	")
	fmt.Println(string(b), err)

	if err != nil {
		t.Fatal(err)
	}
}

func TestBuilder_Relation2(t *testing.T) {
	var moments = make([]*MomentList, 0)
	err := Model(&moments).Relation("User"  , func(b *Builder) {
		b.Where("gender = 0")
	}).Where("status = 1").Limit(10).All()

	b , _ :=json.MarshalIndent(moments,"","	")
	fmt.Println(string(b), err)

	if err != nil {
		t.Fatal(err)
	}
}