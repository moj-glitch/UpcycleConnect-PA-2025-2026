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

            $response = api_request(API_URL . "/api/v1/applications", 'GET', array(api_bearer_header()));
            $applications = json_decode($response['body'], true);
        ?>
        <title><?php echo $lang[$ADMIN_APPLICATIONS_TITLE]?></title>
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
                <h1><?php echo $lang[$ADMIN_APPLICATIONS_TITLE]?></h1>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$ANNONCES_NAME_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_PRICE_LABEL]?></th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($applications)): foreach ($applications as $application): ?>
                        <tr>
                            <td><?php echo htmlspecialchars($application['nom']); ?></td>
                            <td>
                                <form action="../private/do.modifierPrixApplication.php" method="POST">
                                    <input type="hidden" name="id" value="<?php echo $application['application_id']; ?>">
                                    <input type="number" step="0.01" min="0" name="prix" value="<?php echo htmlspecialchars($application['prix']); ?>" required>
                                    <button type="submit"><?php echo $lang[$ADMIN_APPLICATIONS_SAVE_LABEL]?></button>
                                </form>
                            </td>
                            <td></td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
