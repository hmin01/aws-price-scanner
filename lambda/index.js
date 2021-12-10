const childProcess = require('child_process');
const path = require('path');

exports.handler = async (event) => {
  // Get a various for process (from process environment and parent lambda function)
  const bucket = process.env.BUCKET;
  const outputDir = process.env.DIRECTORY;
  const serviceCode = event.serviceCode;

  // Execute process
  const result = childProcess.execFileSync(path.join(__dirname, 'priceScanner'), ['-bucket', bucket, '-directory', outputDir, '-srv', serviceCode]);
  const resultStr = result.toString();
  if (resultStr.includes('[ERROR]')) {
    console.error(resultStr);
    return {
      statusCode: 500,
      body: JSON.stringify(resultStr)
    };
  } else {
    return {
      statusCode: 200,
      body: JSON.stringify('Finished')
    };
  }
}