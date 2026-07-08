<html>
    <head>
        <link rel="stylesheet" href="/style.css">
        <meta charset="UTF-8">
        <?php
            require_once __DIR__ . "/private/do.getLanguages.php";
            $lang = $LANGUAGE_CONTENTS;
        ?>
        <title><?= $lang[$INSCRIPTION_TITLE]; ?></title>
    </head>
    <body>
        <header>
            <a href="languages.php?lang=<?= $LOADED_LANGUAGE ?>&redirect=<?= urlencode(basename($_SERVER['PHP_SELF'])); ?>">
                <img
                    src="<?= "./private/lang/" . $LOADED_LANGUAGE . ".svg" ?>"
                    alt=<?= $LOADED_LANGUAGE . " language switch button" ?>
                    height="87"
                    width="100"/>
            </a>
        </header>
        <main>
            <section>
                <?php if (isset($_GET['error'])): ?>
                <p><?php echo $lang[$INSCRIPTION_ERROR_LABEL]; ?></p>
                <?php endif; ?>
                <form action="private/do.inscription.php?lang=<?= $LOADED_LANGUAGE ?>" method="POST">
                    <label for="client_email"><?= $lang[$EMAIL_LABEL]; ?></label>
                    <br>
                    <input type="email" name="client_email" id="client_email" required>
                    <br>
                    <label for="client_secret"><?= $lang[$SECRET_LABEL]; ?></label>
                    <br>
                    <input type="password" name="client_secret" id="client_secret" required>
                    <br>
                    <label for="client_confirm"><?= $lang[$CONFIRM_LABEL]; ?></label>
                    <br>
                    <input type="password" name="client_confirm" id="client_confirm" required>
                    <br>
                    <label for="client_nom"><?= $lang[$INSCRIPTION_NOM_LABEL]; ?></label>
                    <br>
                    <input type="text" name="client_nom" id="client_nom" required>
                    <br>
                    <label for="client_prenom"><?= $lang[$INSCRIPTION_PRENOM_LABEL]; ?></label>
                    <br>
                    <input type="text" name="client_prenom" id="client_prenom" required>
                    <br>
                    <label for="client_telephone"><?= $lang[$INSCRIPTION_TELEPHONE_LABEL]; ?></label>
                    <br>
                    <input type="text" name="client_telephone" id="client_telephone" placeholder="+33612345678" required>
                    <br>
                    <label for="client_adresse"><?= $lang[$INSCRIPTION_ADRESSE_LABEL]; ?></label>
                    <br>
                    <input type="text" name="client_adresse" id="client_adresse" required>
                    <br>
                    <label for="client_code_postal"><?= $lang[$INSCRIPTION_CODE_POSTAL_LABEL]; ?></label>
                    <br>
                    <input type="text" name="client_code_postal" id="client_code_postal" required>
                    <br>
                    <label for="client_ville"><?= $lang[$INSCRIPTION_VILLE_LABEL]; ?></label>
                    <br>
                    <input type="text" name="client_ville" id="client_ville" required>
                    <br>
                    <label for="client_siret"><?= $lang[$INSCRIPTION_SIRET_LABEL]; ?></label>
                    <br>
                    <input type="text" name="client_siret" id="client_siret" required>
                    <br>
                    <label for="account_type"><?= $lang[$INSCRIPTION_TYPE_LABEL]; ?></label>
                    <br>
                    <select name="account_type" id="account_type" onchange="document.getElementById('pro_fields').hidden = this.value != 'entreprise'">
                        <option value="freemium"><?= $lang[$INSCRIPTION_TYPE_PARTICULIER]; ?></option>
                        <option value="entreprise"><?= $lang[$INSCRIPTION_TYPE_PRO]; ?></option>
                    </select>
                    <div id="pro_fields" hidden>
                        <label for="forfait"><?= $lang[$INSCRIPTION_FORFAIT_LABEL]; ?></label>
                        <br>
                        <select name="forfait" id="forfait">
                            <option value="gratuit"><?= $lang[$INSCRIPTION_FORFAIT_GRATUIT]; ?></option>
                            <option value="payant"><?= $lang[$INSCRIPTION_FORFAIT_PAYANT]; ?></option>
                        </select>
                        <br>
                        <label for="entreprise_nom"><?= $lang[$INSCRIPTION_ENTREPRISE_NOM_LABEL]; ?></label>
                        <br>
                        <input type="text" name="entreprise_nom" id="entreprise_nom">
                        <br>
                        <label for="entreprise_adresse"><?= $lang[$INSCRIPTION_ENTREPRISE_ADRESSE_LABEL]; ?></label>
                        <br>
                        <input type="text" name="entreprise_adresse" id="entreprise_adresse">
                        <br>
                        <label for="entreprise_code_postal"><?= $lang[$INSCRIPTION_ENTREPRISE_CP_LABEL]; ?></label>
                        <br>
                        <input type="text" name="entreprise_code_postal" id="entreprise_code_postal">
                        <br>
                        <label for="entreprise_ville"><?= $lang[$INSCRIPTION_ENTREPRISE_VILLE_LABEL]; ?></label>
                        <br>
                        <input type="text" name="entreprise_ville" id="entreprise_ville">
                        <br>
                        <label for="entreprise_siret"><?= $lang[$INSCRIPTION_ENTREPRISE_SIRET_LABEL]; ?></label>
                        <br>
                        <input type="text" name="entreprise_siret" id="entreprise_siret">
                    </div>
                    <br>
                    <button type="submit"><?= $lang[$INSCRIPTION_TITLE]; ?></button>
                </form>
                <a href="connection.php"><?= $lang[$CONNECTION_TITLE]; ?></a>
            </section>
        </main>
        <footer>
            Tout droits reserve a Cycle Connect Enterprise
        </footer>
    </body>
</html>
