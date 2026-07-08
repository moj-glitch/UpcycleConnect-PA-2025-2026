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

            $response = api_request(API_URL . "/api/v1/contrats/categories?from=0&size=100", 'GET', array(api_bearer_header()));
            $data = json_decode($response['body'], true);
            $categories = isset($data['categories']) ? $data['categories'] : (is_array($data) ? $data : array());

            $tiersResponse = api_request(API_URL . "/api/v1/tiers?from=0&size=100", 'GET', array(api_bearer_header()));
            $tiersData = json_decode($tiersResponse['body'], true);
            $tiersList = isset($tiersData['tiers']) ? $tiersData['tiers'] : array();
        ?>
        <title><?php echo $lang[$CONTRATS_NEW_LABEL]?></title>
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
                <h1><?php echo $lang[$CONTRATS_NEW_LABEL]?></h1>
                <form action="../private/do.deposerContrat.php" method="POST" enctype="multipart/form-data">
                    <label for="categorie_contrat_id"><?php echo $lang[$ANNONCES_CATEGORY_LABEL]?></label>
                    <br>
                    <select name="categorie_contrat_id" id="categorie_contrat_id">
                        <?php if (!empty($categories)): foreach ($categories as $categorie): ?>
                        <option value="<?php echo $categorie['categorie_contrat_id']; ?>"><?php echo htmlspecialchars($categorie['libelle']); ?></option>
                        <?php endforeach; endif; ?>
                    </select>
                    <br>
                    <label for="description"><?php echo $lang[$ANNONCES_DESCRIPTION_LABEL]?></label>
                    <br>
                    <textarea name="description" id="description" required></textarea>
                    <br>
                    <label for="depense"><?php echo $lang[$CONTRATS_DEPENSE_LABEL]?></label>
                    <br>
                    <input type="number" step="0.01" min="0" name="depense" id="depense" required>
                    <br>
                    <label for="gain"><?php echo $lang[$CONTRATS_GAIN_LABEL]?></label>
                    <br>
                    <input type="number" step="0.01" min="0" name="gain" id="gain" required>
                    <br>
                    <label for="date_fin"><?php echo $lang[$CONTRATS_DATE_FIN_LABEL]?></label>
                    <br>
                    <input type="datetime-local" name="date_fin" id="date_fin">
                    <br>
                    <label for="pdf"><?php echo $lang[$CONTRATS_PDF_LABEL]?></label>
                    <br>
                    <input type="file" name="pdf" id="pdf" accept="application/pdf" required>
                    <br>
                    <label><?php echo $lang[$CONTRATS_TIERS_LABEL]?></label>
                    <br>
                    <?php if (!empty($tiersList)): foreach ($tiersList as $tier): ?>
                    <input type="checkbox" name="tiers[]" id="tier_<?php echo $tier['tiers_id']; ?>" value="<?php echo $tier['tiers_id']; ?>">
                    <label for="tier_<?php echo $tier['tiers_id']; ?>"><?php echo htmlspecialchars($tier['nom']); ?></label>
                    <br>
                    <?php endforeach; endif; ?>
                    <br>
                    <button type="submit"><?php echo $lang[$CONTRATS_SUBMIT_LABEL]?></button>
                </form>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
