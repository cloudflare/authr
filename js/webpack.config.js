module.exports = {
  entry: './src/Authr.js',
  module: {
    loaders: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        loader: 'babel-loader'
      }
    ]
  },
  externals: {
    lodash: 'lodash'
  },
  output: {
    filename: './build/authr.js',
    library: '@cloudflare/authr',
    libraryTarget: 'umd'
  },
  devtool: 'source-map'
};
