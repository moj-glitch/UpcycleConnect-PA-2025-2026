<?php
if (!isset($_GET['id'])) {
    header("Location: annonces.php?limit=10&offset=0");
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
            $response = api_request(API_URL . "/api/v1/annonces?id=" . $_GET['id'], 'GET', $authHeader);
            $annonce = json_decode($response['body'], true);

            $categoriesResponse = api_request(API_URL . "/api/v1/annonces/categories?from=0&size=100", 'GET', $authHeader);
            $categories = json_decode($categoriesResponse['body'], true);
        ?>
        <title><?php echo $lang[$ANNONCE_EDIT_TITLE]?></title>
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
                <h1><?php echo $lang[$ANNONCE_EDIT_TITLE]?></h1>
                <form action="../private/do.modifierAnnonce.php" method="POST">
                    <input type="hidden" name="id" value="<?php echo $_GET['id']; ?>">
                    <input type="hidden" name="taxe" value="<?php echo htmlspecialchars($annonce['taxe']); ?>">
                    <label for="titre"><?php echo $lang[$ANNONCES_NAME_LABEL]?></label>
                    <br>
                    <input type="text" name="titre" id="titre" value="<?php echo htmlspecialchars($annonce['titre']); ?>" required>
                    <br>
                    <label for="categorie"><?php echo $lang[$ANNONCES_CATEGORY_LABEL]?></label>
                    <br>
                    <select name="categorie" id="categorie" required>
                        <?php if (!empty($categories)): foreach ($categories as $categorie): ?>
                        <option value="<?php echo $categorie['categorie_id']; ?>" <?php echo ($categorie['libelle'] == $annonce['categorie']) ? 'selected' : ''; ?>><?php echo htmlspecialchars($categorie['libelle']); ?></option>
                        <?php endforeach; endif; ?>
                    </select>
                    <br>
                    <label for="prix"><?php echo $lang[$ANNONCES_PRICE_LABEL]?></label>
                    <br>
                    <input type="number" step="0.01" min="0" name="prix" id="prix" value="<?php echo htmlspecialchars($annonce['prix']); ?>" required>
                    <br>
                    <label for="description"><?php echo $lang[$ANNONCES_DESCRIPTION_LABEL]?></label>
                    <br>
                    <textarea name="description" id="description" required><?php echo htmlspecialchars($annonce['description']); ?></textarea>
                    <br>
                    <label for="etat"><?php echo $lang[$ANNONCES_STATE_LABEL]?></label>
                    <br>
                    <select name="etat" id="etat" required>
                        <option value="D" <?php echo ($annonce['etat'] == 'D') ? 'selected' : ''; ?>><?php echo $lang[$ANNONCE_ETAT_D]?></option>
                        <option value="V" <?php echo ($annonce['etat'] == 'V') ? 'selected' : ''; ?>><?php echo $lang[$ANNONCE_ETAT_V]?></option>
                    </select>
                    <br>
                    <label for="image"><?php echo $lang[$ANNONCES_IMAGE_LABEL]?></label>
                    <br>
                    <input type="text" name="image" id="image" value="<?php echo htmlspecialchars($annonce['image']); ?>" required>
                    <br>
                    <br>
                    <button type="submit"><?php echo $lang[$ANNONCE_EDIT_LABEL]?></button>
                </form>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
