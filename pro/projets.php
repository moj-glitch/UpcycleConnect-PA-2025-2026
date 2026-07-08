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
                header("Location: projets.php?limit=10&offset=0");
                exit();
            }

            $limit = $_GET['limit'];
            $offset = $_GET['offset'];

            require_once __DIR__ . "/../private/do.getLanguages.php";
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $response = api_request(API_URL . "/api/v1/projets?size=$limit&from=$offset", 'GET', array(api_bearer_header()));
            $data = json_decode($response['body'], true);
            $projets = isset($data['projets']) ? $data['projets'] : (is_array($data) ? $data : array());
        ?>
        <title><?php echo $lang[$PROJETS_TITLE]?></title>
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
                <h1><?php echo $lang[$PROJETS_TITLE]?></h1>
                <?php if (isset($_GET['campagne'])): ?>
                <p><?php echo $lang[$PROJETS_CAMPAGNE_SUCCESS]?></p>
                <?php endif; ?>
                <a href="projet_new.php"><?php echo $lang[$PROJETS_NEW_LABEL]?></a>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$ANNONCES_NAME_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_DATE_LABEL]?></th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($projets)): foreach ($projets as $projet): ?>
                        <tr>
                            <td><?php echo htmlspecialchars($projet['titre']); ?></td>
                            <td><?php echo date('d/m/Y', strtotime($projet['date_creation'])); ?></td>
                            <td>
                                <form action="../private/do.checkoutCampagne.php" method="POST">
                                    <input type="hidden" name="projet_id" value="<?php echo $projet['projet_id']; ?>">
                                    <input type="hidden" name="titre" value="<?php echo htmlspecialchars($projet['titre']); ?>">
                                    <label><?php echo $lang[$PROJETS_CAMPAGNE_MONTANT_LABEL]?></label>
                                    <input type="number" name="montant" min="100" max="500" step="1" value="100" required>
                                    <button type="submit"><?php echo $lang[$PROJETS_CAMPAGNE_LABEL]?></button>
                                </form>
                            </td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="projets.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="projets.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
