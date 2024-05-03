<template>
    <tr class="d-flex align-items-stretch">
        <td>
            <img :src="track.images[2].url" />
        </td>
        <td class="flex-grow-1 align-middle">
            <h5>
                {{ track.name }}
                <br />
                <small class="text-white-50">{{ track.artist }}</small>
            </h5>
        </td>
    </tr>
</template>

<script lang="ts" setup>
    const emit = defineEmits(["voted"]);
    
    export interface Props {
        track: {
            uri: string;
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
    const runtimeConfig = useRuntimeConfig()
</script>