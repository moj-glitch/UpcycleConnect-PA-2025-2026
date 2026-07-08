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
                header("Location: evenements.php?limit=10&offset=0");
                exit();
            }

            $limit = $_GET['limit'];
            $offset = $_GET['offset'];

            require_once __DIR__ . "/../private/do.getLanguages.php";
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $response = api_request(API_URL . "/api/v1/evenements?size=$limit&from=$offset", 'GET', array(api_bearer_header()));
            $evenements = json_decode($response['body'], true);
        ?>
        <title><?php echo $lang[$EVENEMENTS_TITLE]?></title>
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
                <h1><?php echo $lang[$EVENEMENTS_TITLE]?></h1>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$ANNONCES_NAME_LABEL]?></th>
                            <th><?php echo $lang[$EVENEMENTS_STATUT_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_DATE_LABEL]?></th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($evenements)): foreach ($evenements as $evenement): ?>
                        <tr>
                            <td><a href="evenement.php?id=<?php echo $evenement['evenement_id']; ?>"><?php echo htmlspecialchars($evenement['nom']); ?></a></td>
                            <td><?php echo htmlspecialchars($evenement['statut']); ?></td>
                            <td><?php echo date('d/m/Y', strtotime($evenement['date'])); ?></td>
                            <td>
                                <form action="../private/do.rejoindreEvenement.php" method="POST">
                                    <input type="hidden" name="evenement_id" value="<?php echo $evenement['evenement_id']; ?>">
                                    <button type="submit"><?php echo $lang[$PLANNING_JOIN_LABEL]?></button>
                                </form>
                            </td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="evenements.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="evenements.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
