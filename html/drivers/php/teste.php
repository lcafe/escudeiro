<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Squire's Page</title>
    <link href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body {
            background-color: #f8f9fa;
        }
        .container {
            margin-top: 50px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="card-body">
                <h1 class="card-title">Squire's Page</h1>
                <p class="card-text">Welcome to the world of squires, where bravery and chivalry are the order of the day.</p>
                <?php
                    echo "<p class='card-text'>Squires are young noblemen serving as an attendant to a knight before becoming a knight themselves.</p>";
                ?>
                <button id="myButton" class="btn btn-primary">Learn More</button>
            </div>
        </div>
    </div>
    <script src="https://code.jquery.com/jquery-3.5.1.slim.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/@popperjs/core@2.5.4/dist/umd/popper.min.js"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.min.js"></script>
    <script>
        document.getElementById('myButton').addEventListener('click', function() {
            alert('Squires were essential in medieval times, assisting knights and learning the art of combat.');
        });
    </script>
</body>
</html>
