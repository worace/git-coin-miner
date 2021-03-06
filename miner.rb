require "faraday"
require "digest"
require "securerandom"

BASE_URL = ENV["COIN_URL"] || "http://git-coin.herokuapp.com"

class Miner
  def iteration
    @iteration ||= 1
  end
  def target
    @target ||= Faraday.get("#{BASE_URL}/target").body.hex
  end

  def owner
    ["Alan Kay", "Tim Berners-Lee", "Fred Brooks", "Donald Knuth", "Ada Lovelace", "Grace Hopper", "James Golick", "Weirich", "Adele Goldberg", "Dennis Ritchie", "Ezra Zygmuntowicz", "Yukihiro Matsumoto"].sample
  end

  def mine
    if iteration > 1_000_000
      puts "completed 1mil hashes; refreshing target"
      @target = nil
      @iteration = 1
    end
    input = SecureRandom.hex
    if Digest::SHA1.hexdigest(input).hex < target
      resp = Faraday.post("#{BASE_URL}/hash", {:owner => owner, :message => input})
      puts "got a coin #{input}, resp: #{resp.body}"
      @target = nil
    end
    @iteration = @iteration + 1
  end
end

def run
  (1..4).to_a.map do |i|
    Thread.new do
      miner = Miner.new
      while true do
        miner.mine
      end
    end
  end.map(&:join)
end

run
