<template>
    <tr>
        <td>
            <img :src="track.images[2].url" />
        </td>
        <td class="align-middle">
            <h5>
                {{ track.name }}
                <br />
                <small class="text-muted">{{ track.artist }}</small>
            </h5>
        </td>
        <td class="align-middle">
            <button class="btn btn-primary" @click="addToPlaylist" >Add to playlist</button>
        </td>
    </tr>
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
    $fetch("http://localhost:8888/tracks/add", {method: 'POST', body: { uri: `spotify:track:${props.track.id}` }})
}
</script>
