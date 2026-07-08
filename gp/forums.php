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
                header("Location: forums.php?limit=10&offset=0");
                exit();
            }

            $limit = $_GET['limit'];
            $offset = $_GET['offset'];

            require_once __DIR__ . "/../private/do.getLanguages.php";
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $response = api_request(API_URL . "/api/v1/threads?size=$limit&from=$offset", 'GET', array(api_bearer_header()));
            $forums = json_decode($response['body'], true);
        ?>
        <title><?php echo $lang[$FORUMS_TITLE]?></title>
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
                <h1><?php echo $lang[$FORUMS_TITLE]?></h1>
                <a href="forum_new.php"><?php echo $lang[$FORUM_NEW_LABEL]?></a>
                <br>
                <label for="offset"><?php echo $lang[$FORUMS_OFFSET_LABEL]?></label>
                <select name="offset" id="offset" onchange="window.location.href='forums.php?limit=<?= $limit ?>&offset=' + this.value">
                    <option value="10">10</option>
                    <option value="20">20</option>
                    <option value="50">50</option>
                    <option value="100">100</option>
                </select>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$FORUMS_RESOLVED_LABEL]?></th>
                            <th><?php echo $lang[$FORUMS_NAME_LABEL]?></th>
                            <th><?php echo $lang[$FORUMS_PREVIEW_LABEL]?></th>
                            <th><?php echo $lang[$FORUMS_DATE_LABEL]?></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php
                            if (!empty($forums)):
                                foreach ($forums as $forum):
                        ?>
                        <tr>
                            <td><a href="forum.php?id=<?php echo $forum['thread_id']; ?>"><?php echo ($forum['resolu'] == '1' ? $lang[$FORUMS_RESOLVED_YES] : $lang[$FORUMS_RESOLVED_NO]); ?></a></td>
                            <td><?php echo htmlspecialchars($forum['titre']); ?></td>
                            <td><?php echo htmlspecialchars(mb_substr($forum['message'], 0, 100)); ?>...</td>
                            <td><?php echo date('d/m/Y', strtotime($forum['date_creation'])); ?></td>
                        </tr>
                        <?php
                                endforeach;
                            endif;
                        ?>
                    </tbody>
                </table>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="forums.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="forums.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
