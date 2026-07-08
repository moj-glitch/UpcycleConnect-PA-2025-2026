<html>
    <head>
        <link rel="stylesheet" href="/style.css">
        <meta charset="UTF-8">
        <?php
            require_once __DIR__ . "/private/do.getLanguages.php";
            $lang = $LANGUAGE_CONTENTS;
        ?>
        <title><?php echo $lang[$ROLE_SELECT_TITLE]?></title>
    </head>
    <body id="body">
        <header>
            <a href="languages.php?lang=<?= $LOADED_LANGUAGE ?>&redirect=<?= urlencode(basename($_SERVER['PHP_SELF'])); ?>">
                <img
                    src="<?= "./private/lang/" . $LOADED_LANGUAGE . ".svg" ?>"
                    alt=<?= $LOADED_LANGUAGE . " language switch button" ?>
                    height="87"
                    width="100"/>
            </a>
        </header>
        <main>
            <section>
                <h1><?php echo $lang[$ROLE_SELECT_TITLE]?></h1>
                <h2><?php echo $lang[$ROLE_SELECT_SUBTITLE]?></h2>
                <table>
                    <tbody>
                        <?php
                            $roles = explode(",", $_SESSION['token']['scope']);
                            $applink = $APP_TO_LINK[$_GET['app']];
                            foreach ($roles as $role) {
                                echo "<a href='$applink?role=" . urlencode($role) . "'><tr><td>" . $role . "</td></tr></a>";
                            }
                        ?>
                    </tbody>
                </table>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>