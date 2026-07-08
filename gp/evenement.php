<?php
if (!isset($_GET['id'])) {
    header("Location: evenements.php?limit=10&offset=0");
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

            $response = api_request(API_URL . "/api/v1/evenements?id=" . $_GET['id'], 'GET', array(api_bearer_header()));
            $evenement = json_decode($response['body'], true);
        ?>
        <title><?php echo htmlspecialchars($evenement['nom']); ?></title>
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
                <a href="evenements.php?limit=10&offset=0"><?php echo $lang[$ANNONCE_BACK_LABEL]?></a>
                <h1><?php echo htmlspecialchars($evenement['nom']); ?></h1>
                <p><?php echo $lang[$EVENEMENTS_STATUT_LABEL]?>: <?php echo htmlspecialchars($evenement['statut']); ?></p>
                <p><?php echo $lang[$ANNONCES_DATE_LABEL]?>: <?php echo date('d/m/Y H:i', strtotime($evenement['date'])); ?></p>
                <p><?php echo nl2br(htmlspecialchars($evenement['description'])); ?></p>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
