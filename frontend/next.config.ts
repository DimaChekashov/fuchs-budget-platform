import type { NextConfig } from "next";

const config: NextConfig = {
  async redirects() {
    return [{ source: "/", destination: "/ru", permanent: false }];
  },
};
export default config;
