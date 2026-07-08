<html>
    <head>
        <link rel="stylesheet" href="/style.css">
        <meta charset="UTF-8">
        <?php
            require_once __DIR__ . "/private/do.getLanguages.php";
            $lang = $LANGUAGE_CONTENTS;
        ?>
        <title><?php echo $lang[$LANGUAGES_TITLE]?></title>
        <script>
            document.addEventListener("DOMContentLoaded", function() {
                const languages = <?= json_encode($AVAILABLE_LANGUAGES) ?>;
                let tbody = document.querySelector("table tbody");
                languages.forEach(language => {
                    let tr = document.createElement("tr");
                    let tdFlag = document.createElement("td");
                    let tdCode = document.createElement("td");
                    let flagPath = `./private/lang/${language}.svg`;
                    tdFlag.innerHTML = `<a href="<?=isset($_GET['redirect']) ? $_GET['redirect'] : basename($_SERVER['PHP_SELF']) ?>?lang=${language}"><img src="${flagPath}" alt="${language} flag" height="50" width="60"></a>`;
                    tdCode.innerHTML = `<a href="<?=isset($_GET['redirect']) ? $_GET['redirect'] : basename($_SERVER['PHP_SELF']) ?>?lang=${language}">${language}</a>`;
                    tr.appendChild(tdFlag);
                    tr.appendChild(tdCode);
                    tbody.appendChild(tr);
                    }
                ).catch(error => console.error("Error fetching languages:", error));
            });
        </script>
    </head>
    <body>
        <h1><?php echo $lang[$LANGUAGES_HEADER]?></h1>
        <table>
            <tbody>
            </tbody>
        </table>
    </body>
</html>
