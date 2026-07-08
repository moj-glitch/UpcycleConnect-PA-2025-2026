<html>
    <head>
        <link rel="stylesheet" href="/style.css">
        <meta charset="UTF-8">
        <?php
            session_start();
            if (!isset($_SESSION['token'])) {
                header('Location: connection.php');
                exit();
            }
            require_once __DIR__ . "/private/do.getLanguages.php";
            require_once __DIR__ . "/private/get.appLinks.php";
            require_once __DIR__ . "/api.php";
            $lang = $LANGUAGE_CONTENTS;

            $response = api_request(API_URL . "/api/v1/planning?from=0&size=5&upcoming=1", 'GET', array(api_bearer_header()));
            $upcoming = json_decode($response['body'], true);

            $scoreResponse = api_request(API_URL . "/api/v1/score", 'GET', array(api_bearer_header()));
            $scoreData = json_decode($scoreResponse['body'], true);
        ?>
        <title><?php echo $lang[$MYAPPS_TITLE]?></title>
    </head>
    <body id="body">
        <header>
            <a href="languages.php?lang=<?= $LOADED_LANGUAGE ?>&redirect=<?= urlencode(basename($_SERVER['PHP_SELF'])); ?>">
                <img
                    src="<?= "./private/lang/" . $LOADED_LANGUAGE . ".svg" ?>"
                    alt=<?= $LOADED_LANGUAGE . " language switch button" ?>
                    height="87"
                    width="100"/>
            </a>
        </header>
        <main>
            <section>
                <h1 class="home-title"><?php echo $lang[$MYAPPS_WELCOME]?> <?php echo htmlspecialchars($_SESSION['token']['username']); ?></h1>
                <p><?php echo $lang[$SCORE_LABEL]?>: <?php echo isset($scoreData['score']) ? intval($scoreData['score']) : 0; ?></p>
                <h2><?php echo $lang[$PLANNING_UPCOMING_TITLE]?></h2>
                <table>
                    <tbody>
                        <?php if (!empty($upcoming)): foreach ($upcoming as $entry): ?>
                        <tr>
                            <td><?php echo date('d/m/Y H:i', strtotime($entry['date'])); ?></td>
                            <td><?php echo htmlspecialchars($entry['nom']); ?></td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <a href="<?php echo $APP_TO_LINK['planning']; ?>" target="_blank"><?php echo $lang[$PLANNING_OPEN_LABEL]?></a>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['annonces']; ?>"><?php echo $lang[$MYAPPS_ANNONCES]?></a></td>
                        </tr>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['forums']; ?>"><?php echo $lang[$MYAPPS_FORUMS]?></a></td>
                        </tr>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['evenements']; ?>"><?php echo $lang[$MYAPPS_EVENEMENTS]?></a></td>
                        </tr>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['tutoriels']; ?>"><?php echo $lang[$MYAPPS_TUTORIELS]?></a></td>
                        </tr>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['contrats']; ?>" target="_blank"><?php echo $lang[$MYAPPS_CONTRATS]?></a></td>
                        </tr>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['projets']; ?>" target="_blank"><?php echo $lang[$MYAPPS_PROJETS]?></a></td>
                        </tr>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['relais']; ?>" target="_blank"><?php echo $lang[$MYAPPS_RELAIS]?></a></td>
                        </tr>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['applications']; ?>"><?php echo $lang[$MYAPPS_APPLICATIONS]?></a></td>
                        </tr>
                        <tr>
                            <td><a href="<?php echo $APP_TO_LINK['admin']; ?>"><?php echo $lang[$MYAPPS_ADMIN]?></a></td>
                        </tr>
                    </tbody>
                </table>
                <form action="private/do.deconnexion.php" method="POST">
                    <button type="submit"><?php echo $lang[$MYAPPS_LOGOUT]?></button>
                </form>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
