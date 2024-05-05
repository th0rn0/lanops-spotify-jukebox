<template>
    <div class="container vh-100" style="padding-bottom: 150px;">
        <div class="row h-100">
            <div class="col h-100" style="overflow-y: auto;">
                <h3 class="text-white">Playlist</h3>
                <table class="table table-dark table-striped">
                    <tbody>
                        <PlaylistItem v-for="track in playlist" :track="track" @voted="refreshPlaylist"/>
                    </tbody>
                </table>
            </div>
            <div class="col h-100" style="overflow-y: auto;">
                <h3 class="text-white">Song Search</h3>
                <SongSearch @songAdded="refreshPlaylist" />
            </div>
        </div>
    </div>
    <div v-if="Object.keys(nowPlaying).length !== 0" class="fixed-bottom bg-light p-3">
        <h3>Now Playing</h3>
        <div class="d-flex justify-content-between">
            <div class="d-flex">
                <img :src="nowPlaying.album.images[2].url" />
                <h5 class="ps-3">
                    {{ nowPlaying.name }}
                    <br />
                    <small class="text-muted">{{ nowPlaying.artists[0].name }}</small>
                </h5>
            </div>
            <div class="btn-group-vertical" role="group" aria-label="Basic example">
                <button type="button" class="btn btn-success btn-lg"  @click="voteUp">{{ upVoteText }}</button>
                <button type="button" class="btn btn-danger btn-lg"  @click="voteDown">{{ skipText }}</button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
    const runtimeConfig = useRuntimeConfig()

    const playlist = ref([]);
    const nowPlaying = ref({});
    const upVoteText = ref("Upvote this bop");
    const upVoteShitpost = [
        "Upgrade to LanOps premium™ to unlock this feature",
        "Shit taste, no thanks",
        'Song "upvoted"',
        "Purchase a lanops box and key for a chance to unlock upvotes"
    ]
    const skipText = ref("Vote song uploader is a cunt");
    
    async function refreshPlaylist () {
        playlist.value = await $fetch(runtimeConfig.public.apiEndpoint + `/tracks`).catch((error) => error.data)
    }

    async function voteUp() {
        upVoteText.value =  upVoteShitpost[Math.floor(Math.random() * upVoteShitpost.length)];

    }

    async function voteDown() {
        try {
            await $fetch(`${runtimeConfig.public.apiEndpoint}/votes/skip`, {method: 'POST', body: { uri: nowPlaying.value.uri }})
            skipText.value = "Order received, Obergruppenführer";
            setTimeout(() => skipText.value = "Vote song uploader is a cunt", 10000)
        } catch (error) {
            skipText.value = "Chill out, music hitler";
            setTimeout(() => skipText.value = "Vote song uploader is a cunt", 10000)
        }
        refreshPlaylist();
    }
    
    onMounted(async () => {
        refreshPlaylist();
        nowPlaying.value = await $fetch(runtimeConfig.public.apiEndpoint + `/tracks/current`).catch((error) => error.data)
        setInterval(async () => {
            refreshPlaylist();
            nowPlaying.value = await $fetch(runtimeConfig.public.apiEndpoint + `/tracks/current`).catch((error) => error.data)
        }, 10000);
    })
</script>
