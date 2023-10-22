// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    devtools: { enabled: false },
    runtimeConfig: {
        public: {
            apiEndpoint: process.env.API_ENDPOINT,
        },
    },
    css: [
        '~/assets/scss/main.scss'
    ],
});