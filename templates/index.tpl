<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Video Streaming</title>
</head>

<body>
    <h1>Video Streaming</h1>

    <video id="video" width="900" height="600" controls>
        <!-- No source initially -->
        <source src="http://localhost:3003/v1/stream/play/{{.Filename}}?noConversion={{.NoConversion}}&crf={{.Crf}}">
        Your browser does not support the video tag.
    </video>

</body>
</html>