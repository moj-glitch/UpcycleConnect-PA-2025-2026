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
                header("Location: admin_moderation_annonces.php?limit=10&offset=0");
                exit();
            }

            $limit = $_GET['limit'];
            $offset = $_GET['offset'];

            require_once __DIR__ . "/../private/do.getLanguages.php";
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $response = api_request(API_URL . "/api/v1/annonces?size=$limit&from=$offset", 'GET', array(api_bearer_header()));
            $annonces = json_decode($response['body'], true);
        ?>
        <title><?php echo $lang[$ADMIN_MODERATION_ANNONCES]?></title>
    </head>
    <body id="body">
        <header>
            <a href="admin.php"><?php echo $lang[$ADMIN_TITLE]?></a>
        </header>
        <main>
            <section>
                <h1><?php echo $lang[$ADMIN_MODERATION_ANNONCES]?></h1>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$ANNONCES_NAME_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCE_SELLER_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_PRICE_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_STATE_LABEL]?></th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($annonces)): foreach ($annonces as $annonce): ?>
                        <tr>
                            <td><a href="../gp/annonce.php?id=<?php echo $annonce['annonce_id']; ?>" target="_blank"><?php echo htmlspecialchars($annonce['titre']); ?></a></td>
                            <td><?php echo htmlspecialchars($annonce['vendeur']); ?></td>
                            <td><?php echo number_format($annonce['prix'], 2, ',', ' '); ?>€</td>
                            <td><?php echo htmlspecialchars($annonce['etat']); ?></td>
                            <td>
                                <form action="../private/do.supprimerAnnonceAdmin.php" method="POST">
                                    <input type="hidden" name="id" value="<?php echo $annonce['annonce_id']; ?>">
                                    <button type="submit"><?php echo $lang[$ADMIN_DELETE_LABEL]?></button>
                                </form>
                            </td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="admin_moderation_annonces.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="admin_moderation_annonces.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
