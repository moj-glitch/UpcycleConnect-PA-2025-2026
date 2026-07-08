<?php
if (!isset($_GET['id'])) {
    header("Location: annonces.php?limit=10&offset=0");
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

            $response = api_request(API_URL . "/api/v1/annonces?id=" . $_GET['id'], 'GET', array(api_bearer_header()));
            $annonce = json_decode($response['body'], true);
        ?>
        <title><?php echo htmlspecialchars($annonce['titre']); ?></title>
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
                <a href="annonces.php?limit=10&offset=0"><?php echo $lang[$ANNONCE_BACK_LABEL]?></a>
                <br/>
                <br/>
                <img src="<?php echo htmlspecialchars($annonce['image']); ?>" alt="<?php echo htmlspecialchars($annonce['titre']); ?>">
                <h1><?php echo htmlspecialchars($annonce['titre']); ?></h1>
                <p><?php echo $lang[$ANNONCES_PRICE_LABEL]?>: <?php echo number_format($annonce['prix'], 2, ',', ' '); ?>€</p>
                <p><?php echo $lang[$ANNONCES_CATEGORY_LABEL]?>: <?php echo htmlspecialchars($annonce['categorie']); ?></p>
                <p><?php echo $lang[$ANNONCES_STATE_LABEL]?>: <?php echo htmlspecialchars($annonce['etat']); ?></p>
                <p><?php echo $lang[$ANNONCE_SELLER_LABEL]?>: <?php echo htmlspecialchars($annonce['vendeur']); ?></p>
                <?php if (!empty($annonce['acheteur'])): ?>
                <p><?php echo $lang[$ANNONCE_BUYER_LABEL]?>: <?php echo htmlspecialchars($annonce['acheteur']); ?></p>
                <?php endif; ?>
                <p><?php echo $lang[$ANNONCES_DATE_LABEL]?>: <?php echo date('d/m/Y H:i', strtotime($annonce['date_publication'])); ?></p>
                <h2><?php echo $lang[$ANNONCES_DESCRIPTION_LABEL]?></h2>
                <p><?php echo nl2br(htmlspecialchars($annonce['description'])); ?></p>
                <?php if (!empty($annonce['materiaux'])): ?>
                <h2><?php echo $lang[$ANNONCE_MATERIAUX_LABEL]?></h2>
                <ul>
                    <?php foreach ($annonce['materiaux'] as $materiau): ?>
                    <li><?php echo htmlspecialchars($materiau['nom']); ?></li>
                    <?php endforeach; ?>
                </ul>
                <?php endif; ?>
                <?php if (isset($_GET['error'])): ?>
                <p><?php echo $lang[$ANNONCE_BUY_ERROR_LABEL]?></p>
                <?php endif; ?>
                <?php if ($annonce['etat'] == 'D'): ?>
                <form action="../private/do.acheterAnnonce.php" method="POST">
                    <input type="hidden" name="id" value="<?php echo $_GET['id']; ?>">
                    <button type="submit"><?php echo $lang[$ANNONCE_BUY_LABEL]?></button>
                </form>
                <?php endif; ?>
                <form action="../private/do.supprimerAnnonce.php" method="POST">
                    <input type="hidden" name="id" value="<?php echo $_GET['id']; ?>">
                    <button type="submit"><?php echo $lang[$ANNONCE_DELETE_LABEL]?></button>
                </form>
                <a href="annonce_edit.php?id=<?php echo $_GET['id']; ?>"><?php echo $lang[$ANNONCE_EDIT_LABEL]?></a>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
