<template>
    <div class="container vh-100" style="padding-bottom: 150px;">
        <div class="row h-100">
            <div class="col h-100" style="overflow-y: auto;">
                <h3>Playlist</h3>
                <table class="table table-striped">
                    <tbody>
                        <PlaylistItem v-for="track in playlist" :track="track"/>
                    </tbody>
                </table>
            </div>
            <div class="col h-100" style="overflow-y: auto;">
                <h3>Song Search</h3>
                <SongSearch />
            </div>
        </div>
    </div>
    <div v-if="Object.keys(nowPlaying).length !== 0" class="fixed-bottom bg-light p-3">
        <h3>Now Playing</h3>
        <div class="d-flex">
            <img :src="nowPlaying.album.images[2].url" />
            <h5 class="ps-3">
                {{ nowPlaying.name }}
                <br />
                <small class="text-muted">{{ nowPlaying.artists[0].name }}</small>
            </h5>
        </div>
    </div>
</template>

<script lang="ts" setup>
    const runtimeConfig = useRuntimeConfig()

    const playlist = ref([]);

    onMounted(async () => {
        
    });

    const nowPlaying = ref({})
    
    onMounted(async () => {
        playlist.value = await $fetch(runtimeConfig.public.apiEndpoint + `/tracks`).catch((error) => error.data)
        nowPlaying.value = await $fetch(runtimeConfig.public.apiEndpoint + `/tracks/current`).catch((error) => error.data)
        setInterval(async () => {
            playlist.value = await $fetch(runtimeConfig.public.apiEndpoint + `/tracks`).catch((error) => error.data)
            nowPlaying.value = await $fetch(runtimeConfig.public.apiEndpoint + `/tracks/current`).catch((error) => error.data)
        }, 10000);
    })
</script>
