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
                header("Location: admin_clients.php?limit=10&offset=0");
                exit();
            }

            $limit = $_GET['limit'];
            $offset = $_GET['offset'];
            $q = isset($_GET['q']) ? $_GET['q'] : '';

            require_once __DIR__ . "/../private/do.getLanguages.php";
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $url = API_URL . "/api/v1/admin/clients?size=$limit&from=$offset";
            if ($q != '') {
                $url .= "&q=" . urlencode($q);
            }
            $response = api_request($url, 'GET', array(api_bearer_header()));
            $clients = json_decode($response['body'], true);

            $rolesResponse = api_request(API_URL . "/api/v1/admin/roles", 'GET', array(api_bearer_header()));
            $roles = json_decode($rolesResponse['body'], true);
        ?>
        <title><?php echo $lang[$ADMIN_CLIENTS]?></title>
    </head>
    <body id="body">
        <header>
            <a href="admin.php"><?php echo $lang[$ADMIN_TITLE]?></a>
        </header>
        <main>
            <section>
                <h1><?php echo $lang[$ADMIN_CLIENTS]?></h1>
                <form method="GET" action="admin_clients.php">
                    <input type="hidden" name="limit" value="<?php echo htmlspecialchars($limit); ?>">
                    <input type="hidden" name="offset" value="0">
                    <label for="q"><?php echo $lang[$ADMIN_CLIENTS_SEARCH_LABEL]?></label>
                    <input type="text" name="q" id="q" value="<?php echo htmlspecialchars($q); ?>">
                    <button type="submit"><?php echo $lang[$CONTRATS_SEARCH_SUBMIT]?></button>
                </form>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$EMAIL_LABEL]?></th>
                            <th><?php echo $lang[$INSCRIPTION_NOM_LABEL]?></th>
                            <th><?php echo $lang[$SCORE_LABEL]?></th>
                            <th><?php echo $lang[$ADMIN_CLIENTS_ROLES_LABEL]?></th>
                            <th></th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($clients)): foreach ($clients as $client): ?>
                        <tr>
                            <td><?php echo htmlspecialchars($client['email']); ?></td>
                            <td><?php echo htmlspecialchars($client['prenom']); ?> <?php echo htmlspecialchars($client['nom']); ?></td>
                            <td><?php echo intval($client['score']); ?></td>
                            <td><?php echo htmlspecialchars($client['roles']); ?></td>
                            <td>
                                <form action="../private/do.attribuerRole.php" method="POST">
                                    <input type="hidden" name="client_id" value="<?php echo $client['client_id']; ?>">
                                    <select name="libelle">
                                        <?php if (!empty($roles)): foreach ($roles as $role): ?>
                                        <option value="<?php echo htmlspecialchars($role['libelle']); ?>"><?php echo htmlspecialchars($role['libelle']); ?></option>
                                        <?php endforeach; endif; ?>
                                    </select>
                                    <button type="submit"><?php echo $lang[$ADMIN_CLIENTS_GRANT_LABEL]?></button>
                                </form>
                            </td>
                            <td>
                                <form action="../private/do.retirerRole.php" method="POST">
                                    <input type="hidden" name="client_id" value="<?php echo $client['client_id']; ?>">
                                    <select name="libelle">
                                        <?php if (!empty($roles)): foreach ($roles as $role): ?>
                                        <option value="<?php echo htmlspecialchars($role['libelle']); ?>"><?php echo htmlspecialchars($role['libelle']); ?></option>
                                        <?php endforeach; endif; ?>
                                    </select>
                                    <button type="submit"><?php echo $lang[$ADMIN_CLIENTS_REVOKE_LABEL]?></button>
                                </form>
                            </td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="admin_clients.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>&q=<?= urlencode($q) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="admin_clients.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>&q=<?= urlencode($q) ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
