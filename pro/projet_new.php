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

            $response = api_request(API_URL . "/api/v1/projets/categories?from=0&size=100", 'GET', array(api_bearer_header()));
            $data = json_decode($response['body'], true);
            $categories = isset($data['categories']) ? $data['categories'] : (is_array($data) ? $data : array());
        ?>
        <title><?php echo $lang[$PROJETS_NEW_LABEL]?></title>
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
                <h1><?php echo $lang[$PROJETS_NEW_LABEL]?></h1>
                <form action="../private/do.deposerProjet.php" method="POST" enctype="multipart/form-data">
                    <label for="categorie_projet"><?php echo $lang[$ANNONCES_CATEGORY_LABEL]?></label>
                    <br>
                    <select name="categorie_projet" id="categorie_projet">
                        <?php if (!empty($categories)): foreach ($categories as $categorie): ?>
                        <option value="<?php echo $categorie['categorie_projet_id']; ?>"><?php echo htmlspecialchars($categorie['libelle']); ?></option>
                        <?php endforeach; endif; ?>
                    </select>
                    <br>
                    <label for="titre"><?php echo $lang[$ANNONCES_NAME_LABEL]?></label>
                    <br>
                    <input type="text" name="titre" id="titre" required>
                    <br>
                    <label for="texte"><?php echo $lang[$PROJETS_TEXTE_LABEL]?></label>
                    <br>
                    <textarea name="texte" id="texte" required></textarea>
                    <br>
                    <label for="image"><?php echo $lang[$ANNONCES_IMAGE_LABEL]?></label>
                    <br>
                    <input type="file" name="image" id="image" accept="image/*">
                    <br>
                    <br>
                    <button type="submit"><?php echo $lang[$PROJETS_SUBMIT_LABEL]?></button>
                </form>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
