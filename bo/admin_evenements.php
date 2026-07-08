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
                header("Location: admin_evenements.php?limit=10&offset=0");
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
        <title><?php echo $lang[$ADMIN_EVENEMENTS]?></title>
    </head>
    <body id="body">
        <header>
            <a href="admin.php"><?php echo $lang[$ADMIN_TITLE]?></a>
        </header>
        <main>
            <section>
                <h1><?php echo $lang[$ADMIN_EVENEMENTS]?></h1>
                <h2><?php echo $lang[$ADMIN_EVENEMENT_NEW_LABEL]?></h2>
                <form action="../private/do.deposerEvenement.php" method="POST">
                    <label for="nom"><?php echo $lang[$ANNONCES_NAME_LABEL]?></label>
                    <br>
                    <input type="text" name="nom" id="nom" required>
                    <br>
                    <label for="description"><?php echo $lang[$ANNONCES_DESCRIPTION_LABEL]?></label>
                    <br>
                    <textarea name="description" id="description" required></textarea>
                    <br>
                    <label for="date"><?php echo $lang[$ANNONCES_DATE_LABEL]?></label>
                    <br>
                    <input type="datetime-local" name="date" id="date" required>
                    <br>
                    <label for="statut"><?php echo $lang[$EVENEMENTS_STATUT_LABEL]?></label>
                    <br>
                    <select name="statut" id="statut" required>
                        <option value="P"><?php echo $lang[$ADMIN_EVENEMENT_STATUT_P]?></option>
                        <option value="C"><?php echo $lang[$ADMIN_EVENEMENT_STATUT_C]?></option>
                        <option value="A"><?php echo $lang[$ADMIN_EVENEMENT_STATUT_A]?></option>
                    </select>
                    <br>
                    <label for="categorie"><?php echo $lang[$ANNONCES_CATEGORY_LABEL]?></label>
                    <br>
                    <input type="number" name="categorie" id="categorie" required>
                    <br>
                    <br>
                    <button type="submit"><?php echo $lang[$ADMIN_SUBMIT_LABEL]?></button>
                </form>
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
                            <td><?php echo htmlspecialchars($evenement['nom']); ?></td>
                            <td><?php echo htmlspecialchars($evenement['statut']); ?></td>
                            <td><?php echo date('d/m/Y H:i', strtotime($evenement['date'])); ?></td>
                            <td>
                                <form action="../private/do.supprimerEvenement.php" method="POST">
                                    <input type="hidden" name="id" value="<?php echo $evenement['evenement_id']; ?>">
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
                            <td><a href="admin_evenements.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="admin_evenements.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
