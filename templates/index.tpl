<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Video Streaming Example</title>
</head>
<body>

<h1>Video Streaming</h1>

<video id="streamVideo" width="600" height="420" controls>
    <!-- No source initially -->
    <source src="http://localhost:3003/v1/stream/play/{{.Filename}}">
    Your browser does not support the video tag.
</video>

</body>
</html>