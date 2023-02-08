package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var wg sync.WaitGroup

//Create semaphore for 5 execute downloads synchronously
var sem = make(chan struct{}, 5)

func main() {
	//Abrir a lista de Animes para baixar
	animelist, err := os.Open("Animelist.txt")
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo:", err)
		return
	}
	defer animelist.Close()

	//fmt.Println(animelist)

	//Lear a Lista de Animes
	// Lê o conteúdo do arquivo e visita cada link
	links, err := ioutil.ReadAll(animelist)
	if err != nil {
		fmt.Println("Erro ao ler o arquivo:", err)
		return
	}
	//
	for _, link := range strings.Split(string(links), "\n") {
		// Define o URL que será buscado
		url := link //"https://animefire.net/animes/jujutsu-kaisen-dublado-todos-os-episodios"
		parts := strings.Split(url, "/")
		nome := strings.Split(parts[len(parts)-1], "-todos-os-episodios")[0]

		// Realiza a busca no URL
		doc, err := goquery.NewDocument(url)
		if err != nil {
			fmt.Println("Erro ao obter a página:", err)
			return
		}

		// Abre o arquivo para escrita
		file, err := os.Create("links.txt")
		if err != nil {
			fmt.Println("Erro ao criar o arquivo:", err)
			return
		}
		defer file.Close()

		// Filtra os links que contêm a string "tokyo-revengers-seiya-kessen-hen"
		doc.Find("a").Each(func(i int, selection *goquery.Selection) {
			link, _ := selection.Attr("href")
			if strings.Contains(link, nome) && !strings.Contains(link, "todos-os-episodios") {
				//fmt.Println(link)
				file.WriteString(link + "\n")
			}
		})

		// Abre o arquivo para leitura
		file, err = os.Open("links.txt")
		if err != nil {
			fmt.Println("Erro ao abrir o arquivo:", err)
			return
		}
		defer file.Close()

		// Lê o conteúdo do arquivo e visita cada link
		links, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println("Erro ao ler o arquivo:", err)
			return
		}

		for _, link := range strings.Split(string(links), "\n") {
			if link == "" {
				continue
			}
			//fmt.Println("Visiting", link)
			_, err := http.Get(link)
			if err != nil {
				fmt.Println("Erro ao obter a página:", err)
			}
		}

		extractdownlinks()
	}
	wg.Wait()
}

func extractdownlinks() {
	// Abre o arquivo para escrita
	file, err := os.Open("links.txt")
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo:", err)
		return
	}
	defer file.Close()
	// Lê o conteúdo do arquivo link.txt
	scanner := bufio.NewScanner(file)

	// Cria o arquivo download.txt para escrita
	downloadFile, err := os.Create("download.txt")
	if err != nil {
		fmt.Println("Erro ao criar o arquivo download.txt:", err)
		return
	}
	defer downloadFile.Close()

	// Percorre cada linha do arquivo link.txt
	for scanner.Scan() {
		link := scanner.Text()
		//fmt.Println("Visiting", link)

		// Realiza a busca no link
		doc, err := goquery.NewDocument(link)
		if err != nil {
			fmt.Println("Erro ao obter a página:", err)
			continue
		}

		// Filtra os links da estrutura
		doc.Find("li").Each(func(i int, selection *goquery.Selection) {
			link, exists := selection.Find("a").Attr("href")
			if exists && strings.Contains(link, "download") {
				//fmt.Println("Found download link:", link)
				downloadFile.WriteString(link + "\n")
			}
		})
	}
	//
	// Abre o arquivo para escrita
	filedownload, err := os.Open("download.txt")
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo:", err)
		return
	}
	defer filedownload.Close()
	downloadfromlinks()
}

func downloadfromlinks() {
	// Cria o arquivo download.txt para escrita
	downloadFile, err := os.Create("downloadLink.txt")
	if err != nil {
		fmt.Println("Erro ao criar o arquivo download.txt:", err)
		return
	}
	defer downloadFile.Close()

	// Abrir o arquivo "download.txt"
	file, err := os.Open("download.txt")
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo:", err)
		return
	}
	defer file.Close()

	// Criar um scanner para ler as linhas do arquivo
	scanner := bufio.NewScanner(file)
	// Verificar se houve erros ao ler o arquivo
	if err := scanner.Err(); err != nil {
		fmt.Println("Erro ao ler o arquivo:", err)
		return
	}
	// Percorrer cada linha do arquivo
	for scanner.Scan() {
		link := scanner.Text()
		//fmt.Println("Visiting", link)
		// Realiza a busca no link
		doc, err := goquery.NewDocument(link)
		if err != nil {
			fmt.Println("Erro ao obter a página:", err)
			continue
		}
		// Filtra os links da estrutura
		doc.Find("a").Each(func(i int, selection *goquery.Selection) {
			link, exists := selection.Attr("href") //selection.Find("a").Attr("href")
			if exists && strings.Contains(link, "mp4") && !strings.Contains(link, "googlevideo.com") {
				//fmt.Println("Found download link:", link)
				//

				parts := strings.Split(link, "/")
				foldername := strings.Join(parts[5:7], "-")
				filename := strings.Join(parts[5:8], "-")
				episode := strings.Split(filename, "?")[0]

				index := strings.Index(link, "?")
				downloadLink := link[:index]
				downloadFile.WriteString(downloadLink + "\n")
				//Realize o Download dos link
				//wg.Add(1)
				sem <- struct{}{}
				go downloadVideo(downloadLink, episode, foldername)

			}
		})
	}
}

func downloadVideo(links, filename, foldername string) {

	uri := []string{links}
	for _, link := range uri {
		wg.Add(1)
		resp, err := http.Get(link)
		if err != nil {
			fmt.Errorf("Erro ao realizar o GET: %v", err)
		}
		defer resp.Body.Close()
		if _, err := os.Stat(foldername); os.IsNotExist(err) {
			if err := os.Mkdir(foldername, 0755); err != nil {
				fmt.Println("Erro ao criar a pasta:", err)
			}
		}
		//
		file, err := os.Create(filename) //(filename)
		if err != nil {
			fmt.Println("Erro ao criar o arquivo de vídeo:", err)
		}
		defer file.Close()
		// Obtém o tamanho total do arquivo
		c := &counter{totalSize: resp.ContentLength, filename: filename}
		//Realiza o Download
		_, err = io.Copy(file, io.TeeReader(resp.Body, c)) //resp.Body)
		if err != nil {
			fmt.Println("Erro ao copiar o corpo da resposta para o arquivo:", err)
		}
		//
		err = os.Rename(filename, foldername+"/"+filename)
		if err != nil {
			fmt.Println("Erro ao mover o arquivo para a pasta videos:", err)
		}
		fmt.Printf("\nArquivo '%s' foi baixado com sucesso! Tamanho total: %dMB\n", filename, (c.totalSize / 1000 / 1000))
		<-sem
	}
	defer wg.Done()
}

// counter é uma estrutura de dados que conta a quantidade de dados lidos
type counter struct {
	current   int64
	totalSize int64
	filename  string
}

func (c *counter) Write(p []byte) (int, error) {
	n := len(p)
	c.current += int64(n)
	percent := float64(c.current) / float64(c.totalSize) * 100

	fmt.Printf("\rFileName:%s Size:%.0fMB Percet Downloaded: %.2f%% ", c.filename, float64(c.totalSize/1000/1000), percent)
	return n, nil
}
