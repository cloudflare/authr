const UglifyJSPlugin = require('uglifyjs-webpack-plugin');
module.exports = {
  entry: './src/authr/index.js',
  module: {
    loaders: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        loader: 'babel-loader'
      }
    ]
  },
  output: {
    filename: './build/authr.js',
    library: 'authr',
    libraryTarget: 'umd'
  },
  plugins: [
    new UglifyJSPlugin()
  ],
  devtool: 'source-map'
};
