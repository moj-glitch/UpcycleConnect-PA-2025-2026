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

            $response = api_request(API_URL . "/api/v1/annonces/categories?from=0&size=100", 'GET', array(api_bearer_header()));
            $categories = json_decode($response['body'], true);

            $materiauxResponse = api_request(API_URL . "/api/v1/materiaux?densite_min=0&prix_kg_min=0", 'GET', array(api_bearer_header()));
            $materiaux = json_decode($materiauxResponse['body'], true);
        ?>
        <title><?php echo $lang[$ANNONCE_CREATE_TITLE]?></title>
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
                <h1><?php echo $lang[$ANNONCE_CREATE_TITLE]?></h1>
                <form action="../private/do.deposerAnnonce.php" method="POST">
                    <label for="titre"><?php echo $lang[$ANNONCES_NAME_LABEL]?></label>
                    <br>
                    <input type="text" name="titre" id="titre" required>
                    <br>
                    <label for="categorie"><?php echo $lang[$ANNONCES_CATEGORY_LABEL]?></label>
                    <br>
                    <select name="categorie" id="categorie" required>
                        <?php if (!empty($categories)): foreach ($categories as $categorie): ?>
                        <option value="<?php echo $categorie['categorie_id']; ?>"><?php echo htmlspecialchars($categorie['libelle']); ?></option>
                        <?php endforeach; endif; ?>
                    </select>
                    <br>
                    <label for="prix"><?php echo $lang[$ANNONCES_PRICE_LABEL]?></label>
                    <br>
                    <input type="number" step="0.01" min="0" name="prix" id="prix" required>
                    <br>
                    <label for="description"><?php echo $lang[$ANNONCES_DESCRIPTION_LABEL]?></label>
                    <br>
                    <textarea name="description" id="description" required></textarea>
                    <br>
                    <label for="image"><?php echo $lang[$ANNONCES_IMAGE_LABEL]?></label>
                    <br>
                    <input type="text" name="image" id="image" required>
                    <br>
                    <label><?php echo $lang[$ANNONCE_MATERIAUX_LABEL]?></label>
                    <br>
                    <?php if (!empty($materiaux)): foreach ($materiaux as $materiau): ?>
                    <input type="checkbox" name="materiaux[]" id="materiau_<?php echo $materiau['materieau_id']; ?>" value="<?php echo $materiau['materieau_id']; ?>">
                    <label for="materiau_<?php echo $materiau['materieau_id']; ?>"><?php echo htmlspecialchars($materiau['nom']); ?></label>
                    <br>
                    <?php endforeach; endif; ?>
                    <br>
                    <button type="submit"><?php echo $lang[$ANNONCE_CREATE_SUBMIT_LABEL]?></button>
                </form>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
