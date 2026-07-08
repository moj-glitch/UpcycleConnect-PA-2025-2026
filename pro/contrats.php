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
                header("Location: contrats.php?limit=10&offset=0");
                exit();
            }

            $limit = $_GET['limit'];
            $offset = $_GET['offset'];
            $q = isset($_GET['q']) ? $_GET['q'] : '';
            $tiers = isset($_GET['tiers']) ? $_GET['tiers'] : '';

            require_once __DIR__ . "/../private/do.getLanguages.php";
            require_once __DIR__ . "/../api.php";
            $lang = $LANGUAGE_CONTENTS;

            $url = API_URL . "/api/v1/contrats?size=$limit&from=$offset";
            if ($q != '') {
                $url .= "&q=" . urlencode($q);
            }
            if ($tiers != '') {
                $url .= "&tiers=" . urlencode($tiers);
            }
            $response = api_request($url, 'GET', array(api_bearer_header()));
            $data = json_decode($response['body'], true);
            $contrats = isset($data['contrats']) ? $data['contrats'] : array();
        ?>
        <title><?php echo $lang[$CONTRATS_TITLE]?></title>
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
                <h1><?php echo $lang[$CONTRATS_TITLE]?></h1>
                <a href="contrat_new.php"><?php echo $lang[$CONTRATS_NEW_LABEL]?></a>
                <form method="GET" action="contrats.php">
                    <input type="hidden" name="limit" value="<?php echo htmlspecialchars($limit); ?>">
                    <input type="hidden" name="offset" value="0">
                    <label for="numero"><?php echo $lang[$CONTRATS_SEARCH_ID_LABEL]?></label>
                    <input type="number" name="numero" id="numero">
                    <label for="q"><?php echo $lang[$CONTRATS_SEARCH_Q_LABEL]?></label>
                    <input type="text" name="q" id="q" value="<?php echo htmlspecialchars($q); ?>">
                    <label for="tiers"><?php echo $lang[$CONTRATS_SEARCH_TIERS_LABEL]?></label>
                    <input type="text" name="tiers" id="tiers" value="<?php echo htmlspecialchars($tiers); ?>">
                    <button type="submit" onclick="if(document.getElementById('numero').value){window.location.href='contrat.php?id='+document.getElementById('numero').value;return false;}"><?php echo $lang[$CONTRATS_SEARCH_SUBMIT]?></button>
                </form>
                <table>
                    <thead>
                        <tr>
                            <th><?php echo $lang[$CONTRATS_SEARCH_ID_LABEL]?></th>
                            <th><?php echo $lang[$ANNONCES_DESCRIPTION_LABEL]?></th>
                            <th><?php echo $lang[$CONTRATS_DEPENSE_LABEL]?></th>
                            <th><?php echo $lang[$CONTRATS_GAIN_LABEL]?></th>
                            <th><?php echo $lang[$CONTRATS_DATE_DEBUT_LABEL]?></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php if (!empty($contrats)): foreach ($contrats as $contrat): ?>
                        <tr>
                            <td><a href="contrat.php?id=<?php echo $contrat['contrat_id']; ?>" target="_blank"><?php echo $contrat['contrat_id']; ?></a></td>
                            <td><?php echo htmlspecialchars(mb_substr($contrat['description'], 0, 100)); ?></td>
                            <td><?php echo number_format($contrat['depense'], 2, ',', ' '); ?>€</td>
                            <td><?php echo number_format($contrat['gain'], 2, ',', ' '); ?>€</td>
                            <td><?php echo date('d/m/Y', strtotime($contrat['date_debut'])); ?></td>
                        </tr>
                        <?php endforeach; endif; ?>
                    </tbody>
                </table>
                <table>
                    <tbody>
                        <tr>
                            <td><a href="contrats.php?limit=<?= $limit ?>&offset=<?= max(0, $offset - $limit) ?>&q=<?= urlencode($q) ?>&tiers=<?= urlencode($tiers) ?>"><?= $lang[$PREVIOUS_LABEL]?></a></td>
                            <td><a href="contrats.php?limit=<?= $limit ?>&offset=<?= $offset + $limit ?>&q=<?= urlencode($q) ?>&tiers=<?= urlencode($tiers) ?>"><?= $lang[$NEXT_LABEL]?></a></td>
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
