<?php
if (!isset($_GET['id'])) {
    header("Location: contrats.php?limit=10&offset=0");
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

            $response = api_request(API_URL . "/api/v1/contrats?id=" . $_GET['id'], 'GET', array(api_bearer_header()));
            $contrat = json_decode($response['body'], true);

            $tiersResponse = api_request(API_URL . "/api/v1/tiers?from=0&size=100", 'GET', array(api_bearer_header()));
            $tiersData = json_decode($tiersResponse['body'], true);
            $tiersList = isset($tiersData['tiers']) ? $tiersData['tiers'] : array();

            $linkedTiersIds = array();
            if (!empty($contrat['tiers'])) {
                foreach ($contrat['tiers'] as $linked) {
                    $linkedTiersIds[] = $linked['tiers_id'];
                }
            }
        ?>
        <title><?php echo $lang[$CONTRATS_TITLE]?> <?php echo htmlspecialchars($_GET['id']); ?></title>
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
                <a href="contrats.php?limit=10&offset=0"><?php echo $lang[$ANNONCE_BACK_LABEL]?></a>
                <h1><?php echo $lang[$CONTRATS_TITLE]?> <?php echo htmlspecialchars($_GET['id']); ?></h1>
                <form action="../private/do.modifierContrat.php?id=<?php echo urlencode($_GET['id']); ?>" method="POST">
                    <label for="description"><?php echo $lang[$ANNONCES_DESCRIPTION_LABEL]?></label>
                    <br>
                    <textarea name="description" id="description" required><?php echo htmlspecialchars($contrat['description']); ?></textarea>
                    <br>
                    <label for="depense"><?php echo $lang[$CONTRATS_DEPENSE_LABEL]?></label>
                    <br>
                    <input type="number" step="0.01" min="0" name="depense" id="depense" value="<?php echo htmlspecialchars($contrat['depense']); ?>" required>
                    <br>
                    <label for="gain"><?php echo $lang[$CONTRATS_GAIN_LABEL]?></label>
                    <br>
                    <input type="number" step="0.01" min="0" name="gain" id="gain" value="<?php echo htmlspecialchars($contrat['gain']); ?>" required>
                    <br>
                    <label for="date_fin"><?php echo $lang[$CONTRATS_DATE_FIN_LABEL]?></label>
                    <br>
                    <input type="datetime-local" name="date_fin" id="date_fin" value="<?php echo !empty($contrat['date_fin']) ? date('Y-m-d\TH:i', strtotime($contrat['date_fin'])) : ''; ?>">
                    <br>
                    <p><?php echo $lang[$CONTRATS_PDF_LABEL]?>: <?php echo htmlspecialchars($contrat['pdf']); ?></p>
                    <label><?php echo $lang[$CONTRATS_TIERS_LABEL]?></label>
                    <br>
                    <?php if (!empty($tiersList)): foreach ($tiersList as $tier): ?>
                    <input type="checkbox" name="tiers[]" id="tier_<?php echo $tier['tiers_id']; ?>" value="<?php echo $tier['tiers_id']; ?>" <?php echo in_array($tier['tiers_id'], $linkedTiersIds) ? 'checked' : ''; ?>>
                    <label for="tier_<?php echo $tier['tiers_id']; ?>"><?php echo htmlspecialchars($tier['nom']); ?></label>
                    <br>
                    <?php endforeach; endif; ?>
                    <button type="submit"><?php echo $lang[$CONTRATS_SUBMIT_LABEL]?></button>
                </form>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
