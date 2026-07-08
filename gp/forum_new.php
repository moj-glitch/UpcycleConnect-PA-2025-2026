<html>
    <head>
        <link rel="stylesheet" href="/style.css">
        <meta charset="UTF-8">
        <?php
            session_start();
            if (!isset($_SESSION['token'])) {
                header("Location: ../connection.php");
                exit();
            }

            require_once __DIR__ . "/../private/do.getLanguages.php";
            $lang = $LANGUAGE_CONTENTS;
        ?>
        <title><?php echo $lang[$FORUM_CREATE_TITLE]?></title>
    </head>
    <body id="body">
        <header>
            <a href="../languages.php?lang=<?= $LOADED_LANGUAGE ?>&redirect=<?= urlencode(basename($_SERVER['PHP_SELF'])); ?>">
                <img
                    src="<?= "../private/lang/" . $LOADED_LANGUAGE . ".svg" ?>"
                    alt=<?= $LOADED_LANGUAGE . " language switch button" ?>
                    height="87"
                    width="100"/>
            </a>
        </header>
        <main>
            <section>
                <h1><?php echo $lang[$FORUM_CREATE_TITLE]?></h1>
                <form action="../private/do.deposerThread.php" method="POST">
                    <label for="categorie_thread"><?php echo $lang[$FORUM_CREATE_CATEGORY_LABEL]?></label>
                    <br>
                    <input type="number" name="categorie_thread" id="categorie_thread" required>
                    <br>
                    <label for="titre"><?php echo $lang[$FORUMS_NAME_LABEL]?></label>
                    <br>
                    <input type="text" name="titre" id="titre" required>
                    <br>
                    <label for="message"><?php echo $lang[$FORUMS_PREVIEW_LABEL]?></label>
                    <br>
                    <textarea name="message" id="message" required></textarea>
                    <br>
                    <br>
                    <button type="submit"><?php echo $lang[$FORUM_CREATE_SUBMIT_LABEL]?></button>
                </form>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
