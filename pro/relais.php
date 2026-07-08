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

            if (!isset($_GET['limit']) || !isset($_GET['offset'])) {
                header("Location: relais.php?limit=10&offset=0");
                exit();
            }

            $limit = $_GET['limit'];
            $offset = $_GET['offset'];

            require_once __DIR__ . "/../private/do.getLanguages.php";
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $response = api_request(API_URL . "/api/v1/annonces?size=$limit&from=$offset&mes_achats=1", 'GET', array(api_bearer_header()));
            $annonces = json_decode($response['body'], true);
        ?>
        <title><?php echo $lang[$RELAIS_TITLE]?></title>
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
                <h1><?php echo $lang[$RELAIS_TITLE]?></h1>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$RELAIS_BARCODE_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_NAME_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCE_SELLER_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_PRICE_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_STATE_LABEL]?></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($annonces)): foreach ($annonces as $annonce): ?>
                        <tr>
                            <td><?php echo !empty($annonce['barcode']) ? htmlspecialchars($annonce['barcode']) : ''; ?></td>
                            <td><a href="../gp/annonce.php?id=<?php echo $annonce['annonce_id']; ?>" target="_blank"><?php echo htmlspecialchars($annonce['titre']); ?></a></td>
                            <td><?php echo htmlspecialchars($annonce['vendeur']); ?></td>
                            <td><?php echo number_format($annonce['prix'], 2, ',', ' '); ?>€</td>
                            <td><?php echo htmlspecialchars($annonce['etat']); ?></td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="relais.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="relais.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>"><?= $lang[$NEXT_LABEL]?></a></td>
                        </tr>
                    </tbody>
                </table>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
