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
                header("Location: planning.php?limit=10&offset=0");
                exit();
            }

            $limit = $_GET['limit'];
            $offset = $_GET['offset'];

            require_once __DIR__ . "/../private/do.getLanguages.php";
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $response = api_request(API_URL . "/api/v1/planning?size=$limit&from=$offset", 'GET', array(api_bearer_header()));
            $entries = json_decode($response['body'], true);
        ?>
        <title><?php echo $lang[$PLANNING_TITLE]?></title>
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
                <h1><?php echo $lang[$PLANNING_TITLE]?></h1>
                <a href="evenements.php?limit=10&offset=0"><?php echo $lang[$PLANNING_NEW_LABEL]?></a>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$ANNONCES_DATE_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_NAME_LABEL]?></th>
                            <th><?php echo $lang[$EVENEMENTS_STATUT_LABEL]?></th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($entries)): foreach ($entries as $entry): ?>
                        <tr>
                            <td><?php echo date('d/m/Y H:i', strtotime($entry['date'])); ?></td>
                            <td><a href="evenement.php?id=<?php echo $entry['evenement_id']; ?>"><?php echo htmlspecialchars($entry['nom']); ?></a></td>
                            <td><?php echo htmlspecialchars($entry['statut']); ?></td>
                            <td>
                                <form action="../private/do.quitterEvenement.php" method="POST">
                                    <input type="hidden" name="evenement_id" value="<?php echo $entry['evenement_id']; ?>">
                                    <button type="submit"><?php echo $lang[$PLANNING_LEAVE_LABEL]?></button>
                                </form>
                            </td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="planning.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="planning.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
