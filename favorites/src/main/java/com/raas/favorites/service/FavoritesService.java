package com.raas.favorites.service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import org.springframework.data.redis.core.HashOperations;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

@Service
public class FavoritesService {
    private final HashOperations<String, String, String> hashOperations;

    public FavoritesService(StringRedisTemplate redisTemplate) {
        this.hashOperations = redisTemplate.opsForHash();
    }

    public void addFavorite(String userId, String listingId) {
        hashOperations.put(key(userId), listingId, "1");
    }

    public void removeFavorite(String userId, String listingId) {
        hashOperations.delete(key(userId), listingId);
    }

    public List<String> listFavorites(String userId) {
        Map<String, String> entries = hashOperations.entries(key(userId));
        return new ArrayList<>(entries.keySet());
    }

    private String key(String userId) {
        return "user:" + userId + ":favorites";
    }
}
