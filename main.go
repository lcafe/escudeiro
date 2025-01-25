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
        /* 🎨 Paleta de Cores (Claro) */
        :root {
            --bg-color: #f8f9fa; /* Fundo principal */
            --container-bg: #fff; /* Fundo do container */
            --text-color: #000; /* Preto para texto */
            --border-color: #000; /* Bordas escuras */
            --hover-bg: #f1f1f1; /* Cinza claro no hover */
        }

        /* 🌙 Paleta de Cores (Escuro) */
        .dark-mode {
            --bg-color: #121212;
            --container-bg: #1e1e1e;
            --text-color: #ffffff;
            --border-color: #ffffff;
            --hover-bg: #2c2c2c;
        }

        /* 🌐 Reset & Estrutura */
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
            font-family: Arial, sans-serif;
        }

        body {
            background: var(--bg-color);
            color: var(--text-color);
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: flex-start;
            height: 100vh;
            width: 100vw;
            padding: 40px;
            transition: background 0.3s, color 0.3s;
        }

        /* 📂 Container Principal */
        .container {
            width: 100%;
            max-width: 900px;
            background: var(--container-bg);
            border: 2px solid var(--border-color);
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0px 4px 8px rgba(0, 0, 0, 0.1);
            transition: background 0.3s, border 0.3s;
        }

        /* 🏷️ Cabeçalho */
        h2 {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 20px;
        }

        /* 🔙 Botão Voltar */
        .back-link {
            display: flex;
            align-items: center;
            font-size: 18px;
            font-weight: bold;
            text-decoration: none;
            color: var(--text-color);
            margin-bottom: 15px;
        }

        /* 📜 Lista de Arquivos */
        ul {
            list-style: none;
            padding: 0;
            margin: 0;
            width: 100%;
        }

        li {
            display: flex;
            align-items: center;
            padding: 12px;
            border-top: 1px solid var(--border-color);
            transition: background 0.3s ease;
        }

        li:hover {
            background: var(--hover-bg);
        }

        /* 🔗 Links */
        a {
            text-decoration: none;
            color: var(--text-color);
            font-size: 18px;
            display: flex;
            align-items: center;
            width: 100%;
        }

        /* 📂 Ícones SVG */
        .icon {
            width: 24px;
            height: 24px;
            margin-right: 10px;
        }

        /* 🌙 Botão de Troca de Tema */
        .theme-toggle {
            position: fixed;
            top: 10px;
            right: 10px;
            background: none;
            border: 2px solid var(--text-color);
            color: var(--text-color);
            padding: 5px 12px;
            font-size: 14px;
            cursor: pointer;
            border-radius: 4px;
            transition: background 0.3s, color 0.3s;
        }

        .theme-toggle:hover {
            background: var(--text-color);
            color: var(--container-bg);
        }
    </style>
</head>
<body>

    <button class="theme-toggle" onclick="toggleTheme()">🌙 Modo Escuro</button>

    <div class="container">
        <h2>Escudeiro</h2>

        {{if .CurrentPath}}
            <a href="../" class="back-link">
                <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="15 18 9 12 15 6"></polyline>
                </svg>
                Diretório:
            </a>
        {{end}}

        <ul>
            {{range .Files}}
            <li>
                {{if isDir .}}
                    <a href="{{.}}/">
                        <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M3 6h18a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2z"></path>
                            <path d="M3 6l3-3h6l3 3"></path>
                        </svg>
                        {{.}}
                    </a>
                {{else}}
                    <a href="/files/{{$.CurrentPath}}{{.}}" target="_blank">
                        <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M6 2h8l6 6v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2z"></path>
                            <path d="M14 2v6h6"></path>
                        </svg>
                        {{.}}
                    </a>
                {{end}}
            </li>
            {{end}}
        </ul>
    </div>

    <script>
        // Carregar o tema salvo no localStorage
        document.addEventListener("DOMContentLoaded", function() {
            if (localStorage.getItem("theme") === "dark") {
                document.body.classList.add("dark-mode");
                document.querySelector('.theme-toggle').innerText = "☀️ Modo Claro";
            }
        });

        // Alternar tema e salvar no localStorage
        function toggleTheme() {
            document.body.classList.toggle('dark-mode');
            let button = document.querySelector('.theme-toggle');

            if (document.body.classList.contains('dark-mode')) {
                button.innerText = "☀️ Modo Claro";
                localStorage.setItem("theme", "dark");
            } else {
                button.innerText = "🌙 Modo Escuro";
                localStorage.setItem("theme", "light");
            }
        }
    </script>

</body>
</html>

`
