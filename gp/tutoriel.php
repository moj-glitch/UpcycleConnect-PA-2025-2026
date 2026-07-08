<?php
if (!isset($_GET['id'])) {
    header("Location: tutoriels.php?limit=10&offset=0");
    exit();
}
?>
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
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $response = api_request(API_URL . "/api/v1/tutoriels?id=" . $_GET['id'], 'GET', array(api_bearer_header()));
            $tutoriel = json_decode($response['body'], true);
        ?>
        <title><?php echo htmlspecialchars($tutoriel['titre']); ?></title>
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
                <a href="tutoriels.php?limit=10&offset=0"><?php echo $lang[$ANNONCE_BACK_LABEL]?></a>
                <h1><?php echo htmlspecialchars($tutoriel['titre']); ?></h1>
                <p><?php echo date('d/m/Y', strtotime($tutoriel['date_creation'])); ?></p>
                <?php if (!empty($tutoriel['video'])): ?>
                <video src="<?php echo htmlspecialchars($tutoriel['video']); ?>" controls></video>
                <?php endif; ?>
                <p><?php echo nl2br(htmlspecialchars($tutoriel['article'])); ?></p>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
