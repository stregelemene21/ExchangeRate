package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-xmlfmt/xmlfmt"
)

//структуры для парсинга xml файлов

type Valute struct {
	//XMLName  xml.Name `xml:"Valute"`
	//NumCode  int      `xml:"NumCode"`
	//CharCode string   `xml:"CharCode"`
	Nominal string `xml:"Nominal"`
	Name    string `xml:"Name"`
	Value   string `xml:"Value"`
}

type ValCurs struct {
	//XMLName xml.Name `xml:"ValCurs"`
	//Date string
	Valutes []Valute `xml:"Valute"`
}

//вывод пустой строки
func EmptyString() {
	fmt.Println("")
}

//функция для загрузки xml файла с сайта цбр
func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

//вычисляем начало периода
func DateFinish(s string) string {
	start1, _ := time.Parse("02/01/2006", s)
	Finish := start1.Add(-2160 * time.Hour)
	Finish1 := Finish.Format("02/01/2006")
	return Finish1
}

//считаем 90 дней назад от указанной даты и возвращаем список дат
func MakeDateList(startDate string) []string {
	S := make([]string, 90)
	layOut := "02/01/2006"
	t, err := time.Parse(layOut, startDate)
	if err != nil {
		fmt.Println("Некорректная дата\nПерезапустите программу")
		os.Exit(1)
	}
	S[0] = t.Format(layOut)
	for i := 1; i < 90; i++ {
		t1 := t.Add(-24 * time.Hour)
		t = t1
		t2 := t1.Format(layOut)
		S[i] = t2
	}
	return S
}

//функция перевода строки в число с плавающей точкой
func StringsToFloat64(s string) float64 {
	c, _ := strconv.ParseFloat(s, 64)
	return c
}

//функция замены "," на "." для корректного вычисления
func StringConvert(s string) string {
	r := strings.NewReplacer(",", ".")
	return r.Replace(s)
}

func main() {

	//выводим описание и запрашиваем дату отсчета
	EmptyString()
	fmt.Println("Программа возвращает минимальное, максимальное и среднее значение курсов валют за 90 дней, предшествующих указанной дате")
	EmptyString()
	fmt.Print("Введите дату в формате ДД/ММ/ГГГГ: ")
	var startDate string
	fmt.Scanln(&startDate)
	EmptyString()
	fmt.Println("Секунду...")
	finishDate := (DateFinish(startDate))
	EmptyString()

	//список дат, данные по которым будем загружать

	datelist := MakeDateList(startDate)

	// итоговая карта со всеми данными по всем дням

	resultMap := make(map[string]ValCurs)

	//получаем и загружаем данные в итоговую карту resultMap

	for _, i := range datelist {
		fileName := "dataxml.txt"
		URL := fmt.Sprintf("http://www.cbr.ru/scripts/XML_daily_eng.asp?date_req=%s", i)
		err := downloadFile(URL, fileName)
		if err != nil {
			//log.Fatal(err)
			fmt.Println("Error opening file:", err)
		}
		xmlFile, err := os.Open("dataxml.txt")
		if err != nil {
			fmt.Println("Error opening file:", err)
		}

		defer xmlFile.Close()

		XMLdata, _ := ioutil.ReadAll(xmlFile)
		z := StringConvert(string(XMLdata))
		x := xmlfmt.FormatXML(string(z), "", " ")

		var C ValCurs

		t := []byte(x)

		xml.Unmarshal(t[46:], &C)
		resultMap[i] = C
		os.Remove("dataxml.txt")
	}

	// карты с запрашиваемыми значениями

	Minimum := make(map[string]string)
	Maximum := make(map[string]string)
	Average := make(map[string]float64)

	// перебираем основную карту resultMap и заполняем карты Minimum, Maximum и Average

	for k := 0; k < 17; k++ {
		N := make(map[string]string)
		for Date, ValCurs := range resultMap {
			val := ValCurs.Valutes[k].Value
			name := ValCurs.Valutes[k].Name + " " + "Nominal: " + ValCurs.Valutes[k].Nominal + " " + Date
			N[val] = name
		}

		//вычисление средних значений

		Sum := make([]float64, 0)
		for key, _ := range N {
			k1 := strings.TrimSpace(key)
			k2, _ := strconv.ParseFloat(k1, 32)
			Sum = append(Sum, k2)
		}
		summ := 0.0
		for _, v := range Sum {
			summ += v
		}
		var total float64
		total = summ / float64(len(Sum))

		//сортировка по значениям валюты в порядке возрастания

		keys := make([]string, 0, len(N))
		for k := range N {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		//добавляем необходимые данные в карты, которые будут выводится

		Average[N[keys[0]]] = total
		Minimum[N[keys[0]]] = keys[0]
		Maximum[N[keys[len(keys)-1]]] = keys[len(keys)-1]
	}

	//выводим итоговые результаты из Minimum, Maximum и Average

	fmt.Println("--------------------------------------------------------------")
	fmt.Printf("Минимальные значения курсов за период %v - %v: \n", finishDate, startDate)
	fmt.Println("--------------------------------------------------------------")
	time.Sleep(1000 * time.Millisecond)
	for i, j := range Minimum {
		j1 := strings.TrimSpace(j) + " ₽"
		fmt.Println(i, j1)
	}
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("---------------------------------------------------------------")
	fmt.Printf("Максимальные значения курсов за период %v - %v: \n", finishDate, startDate)
	fmt.Println("---------------------------------------------------------------")
	time.Sleep(1000 * time.Millisecond)
	for i, j := range Maximum {
		j1 := strings.TrimSpace(j) + " ₽"
		fmt.Println(i, j1)
	}
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("----------------------------------------------------------")
	fmt.Printf("Средние значения курсов за период %v - %v: \n", finishDate, startDate)
	fmt.Println("----------------------------------------------------------")
	time.Sleep(1000 * time.Millisecond)
	for i, j := range Average {
		b := []byte(i)
		d := string(b[:len(b)-10])
		fmt.Printf("%v %.4f ₽\n", d, j)
	}
}
