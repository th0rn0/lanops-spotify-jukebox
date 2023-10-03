<template>
    <div class="accordion-item">
        <h2 class="accordion-header" id="headingOne">
            <button
                class="accordion-button collapsed"
                type="button"
                data-bs-toggle="collapse"
                :data-bs-target="`#song-${track.id}`"
                aria-expanded="false"
                :aria-controls="track.id"
            >
                {{ track.name }} - {{ track.artist }}
            </button>
        </h2>
        <div
            :id="`song-${track.id}`"
            class="accordion-collapse collapse"
            aria-labelledby="headingOne"
            data-bs-parent="#searchList"
        >
            <div class="accordion-body">
                <img :src="track.images[1].url" />
                <button
                    class="btn btn-primary"
                    @click="addToPlaylist"
                >
                    Add to playlist
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
export interface Props {
    track: {
        id: string;
        name: string;
        artist: string;
        images: {
            height: number;
            width: number;
            url: string;
        }[]
    }
}

const props = defineProps<Props>()

function addToPlaylist() {
    console.log(props.track.id)
    $fetch("http://localhost:8888/tracks/add", {method: 'POST', body: { uri: props.track.id }})
}
</script>
