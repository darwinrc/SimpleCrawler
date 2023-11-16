<template>
  <div class="container">
    <h1>Parser Web Crawler</h1>
    <form @click.prevent="onSubmit">
        <input class="url" type="text" v-model="url" placeholder="URL">
        <input class="button" type="submit" value="Crawl URL" @click="crawl" :disabled="processing">
    </form>
    <div class="results" :ref="setScrollableDivRef">
      {{ result }}
      <ul>
        <li v-for="(value, key) in pages" :key="key">
          {{ key }}
          <ul>
            <li v-for="v in value">
              {{ v }}
            </li>
            </ul>
        </li>
      </ul>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      scrollableDiv: null,
      pages: null,
      result: "",
      socket: null,
      url: "https://crawler-test.com/",
      processing: false,
    }
  },
  mounted() {
    this.$nextTick(() => {
      this.scrollableDiv = this.$refs.scrollableDiv
    });
  },
  methods: {
    setScrollableDivRef(el) {
      this.scrollableDiv = el
    },
    scrollToBottom() {
      if (this.scrollableDiv) {
        this.scrollableDiv.scrollTop = this.scrollableDiv.scrollHeight
      }
    },
    async crawl() {
      this.processing = true
      this.socket = new WebSocket("ws://localhost:5000/ws")

      this.socket.onmessage = (evt) => {

        const jsonData = JSON.parse(evt.data)
        this.result = "Crawling results..."

        console.log(jsonData)
        this.pages = jsonData.pages

        this.$nextTick(() => {
          this.scrollToBottom()
        });
      }

      this.socket.onopen = (evt) => {
        let msg = {url: this.url}
        this.socket.send(JSON.stringify(msg))
      }
    }
  },
}
</script>

<style>
#app {
 font-family: sans-serif;
}

.container {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.results {
  width: 666px;
  height: 333px;
  border: black 1px solid;
  overflow-y: scroll;
}

.button {
  margin-left: 33px;
}

.url {
  width: 333px;
  margin: 33px 0;
}
</style>