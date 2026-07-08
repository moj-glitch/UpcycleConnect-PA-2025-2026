<?php
if (!isset($_GET['id'])) {
    header("Location: forums.php?limit=10&offset=0");
    exit();
}
?>
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

            $authHeader = array(api_bearer_header());
            $response = api_request(API_URL . "/api/v1/threads?id=" . $_GET['id'], 'GET', $authHeader);
            $forum = json_decode($response['body'], true);

            $messagesResponse = api_request(API_URL . "/api/v1/threads/messages?thread_id=" . $_GET['id'], 'GET', $authHeader);
            $messages = json_decode($messagesResponse['body'], true);
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
                <h2><?php echo ($forum['resolu'] == '1' ? $lang[$FORUMS_RESOLVED_YES] : $lang[$FORUMS_RESOLVED_NO]); ?> <?php echo htmlspecialchars($forum['titre']); ?></h2>
                <p><?php echo date('d/m/Y H:i', strtotime($forum['date_creation'])); ?></p>
                <p><?php echo nl2br(htmlspecialchars($forum['message'])); ?></p>
                <?php if ($forum['resolu'] != '1'): ?>
                <form action="../private/do.resoudreThread.php" method="POST">
                    <input type="hidden" name="id" value="<?php echo $_GET['id']; ?>">
                    <input type="hidden" name="categorie_thread" value="<?php echo $forum['categorie_thread']; ?>">
                    <input type="hidden" name="titre" value="<?php echo htmlspecialchars($forum['titre']); ?>">
                    <input type="hidden" name="message" value="<?php echo htmlspecialchars($forum['message']); ?>">
                    <button type="submit"><?php echo $lang[$FORUM_RESOLVE_LABEL]?></button>
                </form>
                <?php endif; ?>
                <table>
                    <tbody>
                        <?php if (!empty($messages)): foreach ($messages as $message): ?>
                        <tr>
                            <td>
                                <?php echo $lang[$ANNONCE_AUTHOR_LABEL]?>: <?php echo htmlspecialchars($message['client_id']); ?><br/>
                                <?php echo htmlspecialchars($message['message']); ?><br/>
                                <?php echo date('d/m/Y H:i', strtotime($message['date_envoi'])); ?>
                                <?php
                                    $repliesResponse = api_request(API_URL . "/api/v1/threads/messages/children?thread_id=" . $_GET['id'] . "&parent_id=" . $message['message_thread_id'], 'GET', $authHeader);
                                    $replies = json_decode($repliesResponse['body'], true);
                                    if (!empty($replies)):
                                ?>
                                <table>
                                    <tbody>
                                        <?php foreach ($replies as $reply): ?>
                                        <tr>
                                            <td><?php echo htmlspecialchars($reply['client_id']); ?><br/><?php echo htmlspecialchars($reply['message']); ?></td>
                                        </tr>
                                        <?php endforeach; ?>
                                    </tbody>
                                </table>
                                <?php endif; ?>
                                <form action="../private/do.envoyerMessageThread.php" method="POST">
                                    <input type="hidden" name="thread_id" value="<?php echo $_GET['id']; ?>">
                                    <input type="hidden" name="parent" value="<?php echo $message['message_thread_id']; ?>">
                                    <input type="text" name="message" required>
                                    <button type="submit"><?php echo $lang[$FORUM_REPLY_SUBMIT_LABEL]?></button>
                                </form>
                            </td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <form action="../private/do.envoyerMessageThread.php" method="POST">
                    <input type="hidden" name="thread_id" value="<?php echo $_GET['id']; ?>">
                    <input type="text" name="message" required>
                    <button type="submit"><?php echo $lang[$FORUM_REPLY_SUBMIT_LABEL]?></button>
                </form>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
