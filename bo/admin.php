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
            $lang = $LANGUAGE_CONTENTS;

            $scopes = isset($_SESSION['token']['scope']) ? $_SESSION['token']['scope'] : '';
            function has_scope($scopes, $scope) {
                return in_array($scope, preg_split('/[ ,]+/', $scopes));
            }
            $tiles = array();
            if (has_scope($scopes, 'public:admin_threads')) {
                $tiles[] = array('admin_moderation_threads.php?limit=10&offset=0', $lang[$ADMIN_MODERATION_THREADS]);
            }
            if (has_scope($scopes, 'public:admin_annonces')) {
                $tiles[] = array('admin_moderation_annonces.php?limit=10&offset=0', $lang[$ADMIN_MODERATION_ANNONCES]);
            }
            if (has_scope($scopes, 'evenements:manager')) {
                $tiles[] = array('admin_evenements.php?limit=10&offset=0', $lang[$ADMIN_EVENEMENTS]);
            }
            if (has_scope($scopes, 'tutorials:content_manager')) {
                $tiles[] = array('admin_tutoriels.php', $lang[$ADMIN_TUTORIELS]);
            }
            if (has_scope($scopes, 'admin:applications')) {
                $tiles[] = array('admin_applications.php', $lang[$MYAPPS_ADMIN_APPLICATIONS]);
            }
            if (has_scope($scopes, 'admin:general')) {
                $tiles[] = array('admin_clients.php?limit=10&offset=0', $lang[$ADMIN_CLIENTS]);
            }
        ?>
        <title><?php echo $lang[$ADMIN_TITLE]?></title>
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
                <h1><?php echo $lang[$ADMIN_TITLE]?></h1>
                <?php if (empty($tiles)): ?>
                <p><?php echo $lang[$ADMIN_NO_ACCESS_LABEL]?></p>
                <?php else: ?>
                <table>
                    <tbody>
                        <?php foreach ($tiles as $tile): ?>
                        <tr>
                            <td><a href="<?php echo $tile[0]; ?>"><?php echo $tile[1]; ?></a></td>
                        </tr>
                        <?php endforeach; ?>
                    </tbody>
                </table>
                <?php endif; ?>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
