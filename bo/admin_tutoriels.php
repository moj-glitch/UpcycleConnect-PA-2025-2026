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
        <title><?php echo $lang[$ADMIN_TUTORIELS]?></title>
    </head>
    <body id="body">
        <header>
            <a href="admin.php"><?php echo $lang[$ADMIN_TITLE]?></a>
        </header>
        <main>
            <section>
                <h1><?php echo $lang[$ADMIN_TUTORIELS]?></h1>
                <h2><?php echo $lang[$ADMIN_TUTORIEL_NEW_LABEL]?></h2>
                <form action="../private/do.deposerTutoriel.php" method="POST" enctype="multipart/form-data">
                    <label for="titre"><?php echo $lang[$ANNONCES_NAME_LABEL]?></label>
                    <br>
                    <input type="text" name="titre" id="titre" required>
                    <br>
                    <label for="article"><?php echo $lang[$ADMIN_TUTORIEL_ARTICLE_LABEL]?></label>
                    <br>
                    <textarea name="article" id="article" required></textarea>
                    <br>
                    <label for="categorie"><?php echo $lang[$ANNONCES_CATEGORY_LABEL]?></label>
                    <br>
                    <input type="number" name="categorie" id="categorie" required>
                    <br>
                    <label for="video"><?php echo $lang[$ADMIN_TUTORIEL_VIDEO_LABEL]?></label>
                    <br>
                    <input type="file" name="video" id="video" accept="video/*" required>
                    <br>
                    <br>
                    <button type="submit"><?php echo $lang[$ADMIN_SUBMIT_LABEL]?></button>
                </form>
                <a href="../gp/tutoriels.php?limit=10&offset=0" target="_blank"><?php echo $lang[$TUTORIELS_TITLE]?></a>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
