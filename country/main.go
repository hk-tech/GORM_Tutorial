package main

import (
    "log"
    "net/http"
    "time"

    "github.com/ant0ine/go-json-rest/rest"
    _ "github.com/go-sql-driver/mysql" // エイリアスでprefixを省略できる
    "github.com/jinzhu/gorm"
)

type Country struct {
    Id        int64     `json:"id"`
	Name      string    `sql:"size:1024" json:"name"`
	City      string    `sql:"size:1024" json:"city"`
    CreatedAt time.Time `json:"createdAt"`
}

type Impl struct {
    DB *gorm.DB
}

func (i *Impl) InitDB() {
    var err error
    // MySQLとの接続。ユーザ名：gorm パスワード：password DB名：country
    i.DB, err = gorm.Open("mysql", "gorm:password@/country?charset=utf8&parseTime=True&loc=Local")
    if err != nil {
        log.Fatalf("Got error when connect database, the error is '%v'", err)
    }
    i.DB.LogMode(true)
}

// DBマイグレーション
func (i *Impl) InitSchema() {
    i.DB.AutoMigrate(&Country{})
}

func main() {

    i := Impl{}
    i.InitDB()
    i.InitSchema()

    api := rest.NewApi()
    api.Use(rest.DefaultDevStack...)
    router, err := rest.MakeRouter(
        rest.Get("/countries", i.GetAllCountries),
        rest.Post("/countries", i.PostCountry),
        rest.Get("/countries/:id", i.GetCountry),
        rest.Put("/countries/:id", i.PutCountry),
        rest.Delete("/countries/:id", i.DeleteCountry),
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("server started.")
    api.SetApp(router)
    log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

// countriesテーブル内のデータ全出力
func (i *Impl) GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
    countries := []Country{}
    i.DB.Find(&countries)
    w.WriteJson(&countries)
}

// パスパラメータ:idの国の該当データを出力
func (i *Impl) GetCountry(w rest.ResponseWriter, r *rest.Request) {
    id := r.PathParam("id")
    country := Country{}
    if i.DB.Find(&country, id).Error != nil {
        rest.NotFound(w, r)
        return
    }
    w.WriteJson(&country)
}

// json形式のデータをPOST {name:国名}
func (i *Impl) PostCountry(w rest.ResponseWriter, r *rest.Request) {
    country := Country{}
    err := r.DecodeJsonPayload(&country)
    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
    }
    err = i.DB.Save(&country).Error
    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteJson(&country)
}

// パスパラメータ:idの国の該当データのNameを変更し出力
func (i *Impl) PutCountry(w rest.ResponseWriter, r *rest.Request) {

    id := r.PathParam("id")
    country := Country{}
    if i.DB.First(&country, id).Error != nil {
        rest.NotFound(w, r)
        return
    }
    updated := Country{}
    if err := r.DecodeJsonPayload(&updated); err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    country.Name = updated.Name

    if err := i.DB.Save(&country).Error; err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteJson(&country)
}

// パスパラメータ:idの国の該当データを削除
func (i *Impl) DeleteCountry(w rest.ResponseWriter, r *rest.Request) {
    id := r.PathParam("id")
    country := Country{}
    if i.DB.First(&country, id).Error != nil {
        rest.NotFound(w, r)
        return
    }
    if err := i.DB.Delete(&country).Error; err != nil {
        rest.Error(w, err.Error(), http.StatusInsufficientStorage)
        return
    }
    w.WriteHeader(http.StatusOK)
}