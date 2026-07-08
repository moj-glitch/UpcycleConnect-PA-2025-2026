<html>
    <head>
        <link rel="stylesheet" href="/style.css">
        <meta charset="UTF-8">
        <?php
            session_start();
            require_once __DIR__ . "/private/do.getLanguages.php";
            $lang = $LANGUAGE_CONTENTS;
        ?>
        <title><?php echo $lang[$CONNECTION_TITLE]?></title>
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
                <?php if (isset($_GET['error'])): ?>
                <p><?php echo $lang[$CONNECTION_ERROR_LABEL];?></p>
                <?php endif; ?>
                <form method="POST" action="private/do.connection.php">
                    <label for="client_id"><?php echo $lang[$EMAIL_LABEL];?></label>
                    <br>
                    <input type="text" name="client_id" id="client_id" required>
                    <br>
                    <label for="client_secret"><?php echo $lang[$SECRET_LABEL];?></label>
                    <br>
                    <input type="password" name="client_secret" id="client_secret" required>
                    <br>
                    <br>
                    <input type="submit" value="<?php echo $lang[$CONNECTION_TITLE];?>">
                </form>
                <a href="inscription.php"><?php echo $lang[$INSCRIPTION_TITLE];?></a>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>