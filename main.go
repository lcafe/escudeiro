package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Erro ao carregar o arquivo .env:", err)
	}

	webRoot := os.Getenv("WEB_ROOT")
	if webRoot == "" {
		log.Fatal("❌ A variável de ambiente WEB_ROOT não está definida.")
	}

	if _, err := os.Stat(webRoot); os.IsNotExist(err) {
		log.Fatalf("❌ O diretório WEB_ROOT (%s) não existe!", webRoot)
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	// Criar roteador
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			renderDirectory(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	}) // Listar diretórios dinamicamente

	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			serveFile(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	}) // Servir arquivos (PHP ou estáticos)

	proxyTarget := os.Getenv("PROXY_TARGET")
	if proxyTarget == "" {
		log.Println("⚠️ Nenhuma variável PROXY_TARGET definida. Proxy será desativado.")
	}

	if proxyTarget != "" {
		mux.HandleFunc("/api/", handleProxy(proxyTarget))
	}

	// Configurar servidor
	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      mux,
		ErrorLog:     log.New(os.Stderr, "log: ", log.LstdFlags),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful Shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Println("⚠️  Recebido sinal de interrupção. Encerrando servidor...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		log.Println("✅ Servidor encerrado com sucesso.")
	}()

	log.Printf("🚀 Servidor rodando na porta :%s", serverPort)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Erro no servidor:", err)
	}
}

func handleProxy(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxyURL, err := url.Parse(target)
		if err != nil {
			http.Error(w, "Erro interno ao configurar proxy", http.StatusInternalServerError)
			return
		}

		// Criar o Reverse Proxy
		proxy := httputil.NewSingleHostReverseProxy(proxyURL)

		// Redirecionar requisições para o backend
		log.Printf("🔄 Encaminhando requisição para: %s%s", target, r.URL.Path)
		proxy.ServeHTTP(w, r)
	}
}

func renderDirectory(w http.ResponseWriter, r *http.Request) {
	webRoot := os.Getenv("WEB_ROOT")

	// Obter caminho correto do diretório
	relativePath := strings.TrimPrefix(r.URL.Path, "/")
	directoryPath := filepath.Join(webRoot, relativePath)

	log.Printf("📂 Listando diretório: %s", directoryPath)

	files, err := listFiles(directoryPath)
	if err != nil {
		log.Printf("❌ Erro ao listar arquivos em %s: %s", directoryPath, err)
		http.Error(w, "Falha ao listar diretório", http.StatusInternalServerError)
		return
	}

	// Criar o template corrigido para gerar links corretos
	tmpl := template.Must(template.New("index").Funcs(template.FuncMap{
		"isDir": func(fileName string) bool {
			fileInfo, err := os.Stat(filepath.Join(directoryPath, fileName))
			return err == nil && fileInfo.IsDir()
		},
	}).Parse(htmlTemplate))

	tmpl.Execute(w, struct {
		CurrentPath string
		Files       []string
	}{CurrentPath: relativePath, Files: files})

	log.Println("📤 Diretório renderizado com sucesso.")
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	webRoot := os.Getenv("WEB_ROOT")

	filePath := filepath.Join(webRoot, strings.TrimPrefix(r.URL.Path, "/files/"))

	log.Printf("📥 Tentando servir arquivo: %s", filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		log.Printf("❌ Arquivo ou diretório não encontrado: %s", filePath)
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		indexPath := filepath.Join(filePath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			log.Printf("📄 Servindo index.html automaticamente: %s", indexPath)
			http.ServeFile(w, r, indexPath)
			return
		} else if _, err := os.Stat(filepath.Join(filePath, "index.php")); err == nil {
			log.Printf("📄 Servindo index.php automaticamente: %s", indexPath)
			executePHP(w, filepath.Join(filePath, "index.php"))
			return
		}
	}

	if filepath.Ext(filePath) == ".php" {
		executePHP(w, filePath)
		return
	} else {
		log.Printf("📤 Servindo arquivo: %s", filePath)
		http.ServeFile(w, r, filePath)
	}
}

func executePHP(w http.ResponseWriter, filePath string) {
	phpPath := filepath.Join("drivers", "php", "php.exe") // Caminho do PHP
	log.Printf("⚡ Executando PHP: %s %s", phpPath, filePath)

	cmd := exec.Command(phpPath, filePath)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("❌ Erro ao executar PHP: %s", err)
		http.Error(w, "Erro ao executar PHP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(output)
	log.Printf("✅ Execução do PHP concluída: %s", filePath)
}

func listFiles(directory string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(directory) // Lê SOMENTE o diretório atual
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		files = append(files, entry.Name()) // Adiciona nome do arquivo/pasta
	}

	return files, nil
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="pt">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Escudeiro</title>
    <style>
        /* 🎨 Paleta de Cores */
        :root {
            --bg-color: #f5f5f5; /* Whitesmoke */
            --text-color: #2c2c2c; /* Cinza quase preto */
            --accent-color: #42a5f5; /* Azul Go */
            --hover-bg: #e0e0e0; /* Cinza claro */
            --border-color: #d1d1d1;
        }

        /* 🌐 Reset & Estrutura */
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
        }

        body {
            background: var(--bg-color);
            color: var(--text-color);
            display: flex;
            justify-content: center;
            align-items: center;
            flex-direction: column;
            height: 100vh;
            padding: 20px;
        }

        /* 📂 Container Principal */
        .container {
            width: 90%;
            max-width: 800px;
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
            padding: 20px;
        }

        /* 🏷️ Cabeçalho */
        h2 {
            font-size: 24px;
            font-weight: 600;
            margin-bottom: 15px;
            border-bottom: 2px solid var(--border-color);
            padding-bottom: 10px;
        }

        /* 📜 Lista de Arquivos */
        ul {
            list-style: none;
            padding: 0;
        }

        li {
            padding: 12px 15px;
            border-bottom: 1px solid var(--border-color);
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        li:last-child {
            border-bottom: none;
        }

        /* 🔗 Links */
        a {
            text-decoration: none;
            color: var(--text-color);
            font-size: 18px;
            font-weight: 500;
            transition: color 0.3s ease;
            display: flex;
            align-items: center;
        }

        a:hover {
            color: var(--accent-color);
        }

        /* 📂 Ícones */
        .icon {
            margin-right: 10px;
            font-size: 18px;
        }

        /* 🔙 Botão Voltar */
        .back-link {
            display: inline-block;
            margin-bottom: 15px;
            font-size: 16px;
            font-weight: 500;
            color: var(--accent-color);
            text-decoration: none;
            transition: opacity 0.3s ease;
        }

        .back-link:hover {
            opacity: 0.7;
        }

    </style>
</head>
<body>

    <div class="container">
        <h2>Navegando em: {{.CurrentPath}}</h2>

        {{if .CurrentPath}}
            <a href="../" class="back-link">⬅ Voltar</a>
        {{end}}

        <ul>
            {{range .Files}}
            <li>
                {{if isDir .}}
                    <a href="{{.}}/">
                        <span class="icon">📁</span> {{.}}
                    </a>
                {{else}}
                    <a href="/files/{{$.CurrentPath}}{{.}}" target="_blank">
                        <span class="icon">📄</span> {{.}}
                    </a>
                {{end}}
            </li>
            {{end}}
        </ul>
    </div>

</body>
</html>
`
