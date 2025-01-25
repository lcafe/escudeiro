# Instruções para Baixar e Instalar PHP

Siga estes passos para baixar e instalar o PHP na pasta especificada:

1. **Baixar PHP:**
  - Vá para o [site oficial do PHP](https://www.php.net/downloads).
  - Na seção "Downloads", clique no link para baixar a versão mais recente do PHP no formato `.tar.gz`.

2. **Extrair o Arquivo TAR.GZ:**
  - Após o download ser concluído, navegue até o arquivo `.tar.gz` baixado.
  - Abra um terminal e execute o seguinte comando para extrair o arquivo:
    ```sh
    tar -xvzf php-*.tar.gz -C escudeiro/drivers/
    ```

3. **Configurar PHP:**
  - Navegue até a pasta extraída do PHP.
  - Renomeie o arquivo `php.ini-development` para `php.ini`:
    ```sh
    mv php.ini-development php.ini
    ```
  - Abra o arquivo `php.ini` em um editor de texto:
    ```sh
    nano php.ini
    ```
  - Faça as alterações de configuração necessárias (por exemplo, habilitando extensões).

4. **Adicionar PHP ao Caminho do Sistema:**
  - Abra um terminal e edite o arquivo `.bashrc` ou `.bash_profile`:
    ```sh
    nano ~/.bashrc
    ```
  - Adicione a seguinte linha para incluir a pasta do PHP no seu PATH:
    ```sh
    export PATH=$PATH:escudeiro/drivers/php
    ```
  - Salve o arquivo e execute o seguinte comando para aplicar as mudanças:
    ```sh
    source ~/.bashrc
    ```

5. **Verificar Instalação:**
  - Abra um terminal.
  - Digite `php -v` e pressione Enter.
  - Você deve ver as informações da versão do PHP, indicando que o PHP está instalado corretamente.

Agora o PHP está instalado e configurado na pasta especificada, e você pode executar seus programas PHP.