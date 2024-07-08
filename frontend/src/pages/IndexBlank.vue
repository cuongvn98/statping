<template>
  <div>
    <Group v-for="group in groups" v-bind:key="group.id" :group="group"/>
      <div>
        <div v-for="service in services" :ref="service.id" v-bind:key="service.id">
          <ServiceBlock :service="service"/>
        </div>
      </div>
  </div>
</template>

<script>
import Api from "@/API";

const Group = () => import(/* webpackChunkName: "index" */ '@/components/Index/Group')
const MessageBlock = () => import(/* webpackChunkName: "index" */ '@/components/Index/MessageBlock')
const ServiceBlock = () => import(/* webpackChunkName: "index" */ '@/components/Service/ServiceBlock')
const MessagesIcon = () => import(/* webpackChunkName: "index" */ '@/components/Index/MessagesIcon')

export default {
  name: 'Index',
  components: {
    ServiceBlock,
    MessageBlock,
    MessagesIcon,
    Group,
  },
  data() {
    return {
      logged_in: false,
    }
  },
  computed: {
    loading_text() {
      if (this.$store.getters.groups.length === 0) {
        return "Loading Groups"
      } else if (this.$store.getters.services.length === 0) {
        return "Loading Services"
      } else if (this.$store.getters.messages == null) {
        return "Loading Announcements"
      }
    },
    loaded() {
      return this.$store.getters.services.length !== 0
    },
    core() {
      return this.$store.getters.core
    },
    messages() {
      return this.$store.getters.messages.filter(m => this.inRange(m) && m.service === 0)
    },
    groups() {
      return this.$store.getters.groupsInOrder
    },
    services() {
      return this.$store.getters.servicesInOrder
    },
    services_no_group() {
      return this.$store.getters.servicesNoGroup
    }
  },
  methods: {
    async checkLogin() {
      const token = this.$cookies.get('statping_auth')
      if (!token) {
        this.$store.commit('setLoggedIn', false)
        return
      }
      try {
        const jwt = await Api.check_token(token)
        this.$store.commit('setAdmin', jwt.admin)
        if (jwt.username) {
          this.$store.commit('setLoggedIn', true)
        }
      } catch (e) {
        console.error(e)
      }
    },
    inRange(message) {
      return this.isBetween(this.now(), message.start_on, message.start_on === message.end_on ? this.maxDate().toISOString() : message.end_on)
    }
  }
}
</script>
