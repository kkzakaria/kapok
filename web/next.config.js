/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  env: {
    KAPOK_API_URL: process.env.KAPOK_API_URL || "http://localhost:8080",
  },
};

module.exports = nextConfig;
