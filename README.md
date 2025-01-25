# Escudeiro

## Introdução

**Escudeiro** é um servidor web minimalista desenvolvido em Go, projetado para listar e servir arquivos estáticos e dinâmicos (PHP), além de atuar como um proxy reverso para redirecionamento de requisições. O objetivo é oferecer uma solução leve e eficiente para hospedar e gerenciar conteúdo web localmente.

## Recursos Principais

- Servidor HTTP responsivo e seguro.
- Listagem dinâmica de diretórios e arquivos.
- Suporte a arquivos PHP com execução integrada.
- Proxy reverso para encaminhamento de requisições para um backend.
- Manipulação de variáveis de ambiente via arquivo `.env`.
- Implementação de **Graceful Shutdown** para um desligamento seguro.

## Instalação e Configuração

### 1. Clonando o repositório

```sh
 git clone https://github.com/seu-usuario/escudeiro.git
 cd escudeiro
```

### 2. Configuração das Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto e defina as configurações necessárias:

```
SERVER_PORT=8080
WEB_ROOT=html
PROXY_TARGET=http://localhost:3000
```

- `SERVER_PORT`: Define a porta onde o servidor será iniciado.
- `WEB_ROOT`: Define o diretório raiz para servir arquivos.
- `PROXY_TARGET`: Define o endereço do backend para encaminhar requisições via proxy reverso.

### 3. Executando o Servidor

Para rodar o servidor localmente, execute:

```sh
 go run main.go
```

Ou compile o binário para melhor desempenho:

```sh
 go build -o escudeiro
 ./escudeiro
```

## Uso

### Listagem de Diretórios

Ao acessar `http://localhost:8080/`, o servidor listará os diretórios e arquivos presentes no diretório configurado em `WEB_ROOT`.

### Serviço de Arquivos

Os arquivos dentro de `WEB_ROOT` podem ser acessados diretamente pelo navegador.

Por exemplo:

- `http://localhost:8080/files/meuarquivo.txt`
- `http://localhost:8080/files/imagem.png`

### Execução de Arquivos PHP

Caso um arquivo `.php` seja acessado, ele será interpretado e executado pelo servidor, desde que um interpretador PHP esteja corretamente configurado no diretório `drivers/php/php.exe`.

### Proxy Reverso

O proxy é ativado se a variável `PROXY_TARGET` estiver definida. Todas as requisições que começam com `/api/` serão redirecionadas para o backend especificado.

Por exemplo, se `PROXY_TARGET=http://localhost:3000`, uma requisição para `http://localhost:8080/api/users` será redirecionada para `http://localhost:3000/api/users`.

## Funcionalidades Internas

### 1. Inicialização do Servidor

- Carrega variáveis do `.env`.
- Verifica se o diretório `WEB_ROOT` existe.
- Configura o roteador HTTP.
- Inicia o servidor com **timeouts seguros**.
- Implementa **Graceful Shutdown** para garantir encerramento adequado.

### 2. Roteamento e Serviço de Arquivos

- `/` - Lista os diretórios de `WEB_ROOT` dinamicamente.
- `/files/*` - Serve arquivos estáticos e dinâmicos (PHP incluído).

### 3. Proxy Reverso

- Encaminha requisições para o backend especificado.
- Utiliza `httputil.ReverseProxy` para manutenção de cabeçalhos.

### 4. Execução de PHP

- Se um arquivo `.php` é solicitado, ele é executado via `exec.Command` e o resultado é retornado ao cliente.

## Tecnologias Utilizadas

- **Go** - Linguagem principal.
- **net/http** - Para servir arquivos e manipular requisições HTTP.
- **html/template** - Para renderização dinâmica de diretórios.
- **os/exec** - Para execução de arquivos PHP.
- **httputil.ReverseProxy** - Para implementação do proxy reverso.
- **github.com/joho/godotenv** - Para gerenciamento de variáveis de ambiente.

## Diagrama de Componentes
![diagrama_de_componentes](https://github.com/user-attachments/assets/08da08b9-a3c9-4e6d-93ea-8b50b842a2e6)

### Descrição do Diagrama:

- Servidor HTTP: Atua como o ponto central que recebe todas as requisições HTTP.
- Gerenciador de Arquivos Estáticos: Serve arquivos estáticos (como HTML, CSS, JS) a partir do diretório especificado.
- Executor de PHP: Processa arquivos PHP para gerar respostas dinâmicas.
- Proxy Reverso: Encaminha determinadas requisições para outro servidor ou serviço, conforme configurado.
- Gerenciador de Configurações: Carrega configurações essenciais para o funcionamento do servidor.
- Gerenciador de Desligamento Gradual: Garante que o servidor finalize corretamente, completando requisições em andamento antes de encerrar.

## Contribuição

Se deseja contribuir com melhorias ou reportar problemas, sinta-se à vontade para abrir um **Pull Request** ou **Issue** no [repositório oficial](https://github.com/seu-usuario/escudeiro).

## Licença

Este projeto está licenciado sob a [Apache License 2.0](LICENSE).

---

🚀 **Escudeiro** foi criado como um primeiro projeto para explorar o desenvolvimento web com Go, mantendo um código limpo, eficiente e escalável.


