<template>
    <table class="table table-striped">
        <tbody>
            <PlaylistItem v-for="track in playlist" :track="track"/>
        </tbody>
    </table>
</template>

<script lang="ts" setup>
    const playlist = ref([]);

    onMounted(async () => {
        playlist.value = await $fetch(`http://localhost:8888/tracks`).catch((error) => error.data)
    });
</script>