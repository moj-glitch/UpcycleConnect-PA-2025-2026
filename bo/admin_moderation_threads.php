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
                header("Location: admin_moderation_threads.php?limit=10&offset=0");
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
        <title><?php echo $lang[$ADMIN_MODERATION_THREADS]?></title>
    </head>
    <body id="body">
        <header>
            <a href="admin.php"><?php echo $lang[$ADMIN_TITLE]?></a>
        </header>
        <main>
            <section>
                <h1><?php echo $lang[$ADMIN_MODERATION_THREADS]?></h1>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$FORUMS_NAME_LABEL]?></th>
                            <th><?php echo $lang[$FORUMS_PREVIEW_LABEL]?></th>
                            <th><?php echo $lang[$FORUMS_DATE_LABEL]?></th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($forums)): foreach ($forums as $forum): ?>
                        <tr>
                            <td><a href="../gp/forum.php?id=<?php echo $forum['thread_id']; ?>" target="_blank"><?php echo htmlspecialchars($forum['titre']); ?></a></td>
                            <td><?php echo htmlspecialchars(mb_substr($forum['message'], 0, 80)); ?></td>
                            <td><?php echo date('d/m/Y', strtotime($forum['date_creation'])); ?></td>
                            <td>
                                <form action="../private/do.supprimerThreadAdmin.php" method="POST">
                                    <input type="hidden" name="id" value="<?php echo $forum['thread_id']; ?>">
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
                            <td><a href="admin_moderation_threads.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="admin_moderation_threads.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
