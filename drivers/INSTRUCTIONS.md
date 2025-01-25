# Instructions to Download and Install PHP

Follow these steps to download and install PHP in the specified folder:

1. **Download PHP:**
  - Go to the [official PHP website](https://www.php.net/downloads).
  - Under the "Downloads" section, click on the link to download the latest version of PHP in `.tar.gz` format.

2. **Extract the TAR.GZ File:**
  - Once the download is complete, navigate to the downloaded `.tar.gz` file.
  - Open a terminal and run the following command to extract the file:
    ```sh
    tar -xvzf php-*.tar.gz -C escudeiro/drivers/
    ```

3. **Configure PHP:**
  - Navigate to the extracted PHP folder.
  - Rename the `php.ini-development` file to `php.ini`:
    ```sh
    mv php.ini-development php.ini
    ```
  - Open the `php.ini` file in a text editor:
    ```sh
    nano php.ini
    ```
  - Make any necessary configuration changes (e.g., enabling extensions).

4. **Add PHP to System Path:**
  - Open a terminal and edit the `.bashrc` or `.bash_profile` file:
    ```sh
    nano ~/.bashrc
    ```
  - Add the following line to include the PHP folder in your PATH:
    ```sh
    export PATH=$PATH:escudeiro/drivers/php
    ```
  - Save the file and run the following command to apply the changes:
    ```sh
    source ~/.bashrc
    ```

5. **Verify Installation:**
  - Open a terminal.
  - Type `php -v` and press Enter.
  - You should see the PHP version information, indicating that PHP is installed correctly.

Now PHP is installed and configured in the specified folder, and you can run your PHP programs.