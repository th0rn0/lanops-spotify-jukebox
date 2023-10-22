<template>
    <table class="table table-striped">
        <tbody>
            <PlaylistItem v-for="track in playlist" :track="track"/>
        </tbody>
    </table>
</template>

<script lang="ts" setup>
    const runtimeConfig = useRuntimeConfig()
    const playlist = ref([]);

    onMounted(async () => {
        playlist.value = await $fetch(runtimeConfig.public.apiEndpoint + `/tracks`).catch((error) => error.data)
    });
</script>